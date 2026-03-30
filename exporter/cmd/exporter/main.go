package main

import (
	"bufio"
	"context"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	configv1 "github.com/openshift/api/config/v1"
	configclientset "github.com/openshift/client-go/config/clientset/versioned"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/rest"

	"exporter/pkg/types"
	"exporter/pkg/utils"
)

const (
	EXTRACTOR_ADDRESS string = "127.0.0.1:3000"
)

var extractorTLSCA string
var extractorTLSServerName string

// GatherRuntimeInfoRequest represents the JSON body for POST requests
type GatherRuntimeInfoRequest struct {
	ContainerIds []string `json:"containerIds"`
}

// gatherRuntimeInfo will trigger a new extraction of runtime info
// and reply with a JSON payload
func gatherRuntimeInfo(w http.ResponseWriter, r *http.Request) {
	var containerIds []string

	switch r.Method {
	case "GET":
		// GET scans all containers (empty containerIds)
		containerIds = []string{}
	case "POST":
		var req GatherRuntimeInfoRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error": "Invalid JSON in request body"}`, http.StatusBadRequest)
			return
		}
		if req.ContainerIds == nil {
			http.Error(w, `{"error": "containerIds field is required"}`, http.StatusBadRequest)
			return
		}
		containerIds = req.ContainerIds
	default:
		http.Error(w, "Method is not supported.", http.StatusNotFound)
		return
	}

	hashParam := r.URL.Query().Get("hash")
	hash := hashParam == "" || hashParam == "true"

	startTime := time.Now()
	dataPath, err := triggerRuntimeInfoExtraction(containerIds)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, err = os.Stat(dataPath)
	if dataPath == "" || os.IsNotExist(err) {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer os.RemoveAll(dataPath)

	payload, err := collectWorkloadPayload(hash, dataPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	duration := time.Since(startTime)
	log.Printf("Extracted data from %s (%d bytes) in %s", dataPath, len(response), duration)
	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}

func triggerRuntimeInfoExtraction(containerIds []string) (string, error) {
	caCert, err := os.ReadFile(extractorTLSCA)
	if err != nil {
		return "", fmt.Errorf("failed to read CA certificate: %w", err)
	}
	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		return "", fmt.Errorf("failed to parse CA certificate")
	}

	tlsConfig := &tls.Config{
		RootCAs:    caCertPool,
		ServerName: extractorTLSServerName,
	}

	conn, err := tls.Dial("tcp", EXTRACTOR_ADDRESS, tlsConfig)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	log.Println("Requesting a new runtime extraction")

	payload := strings.Join(containerIds, ",")
	fmt.Fprintf(conn, "%s", payload)
	if err := conn.CloseWrite(); err != nil {
		return "", fmt.Errorf("failed to close write side: %w", err)
	}

	dataPath, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(dataPath), nil
}

func collectWorkloadPayload(hash bool, dataPath string) (types.NodeRuntimeInfo, error) {
	payload := make(types.NodeRuntimeInfo)

	h := sha256.New()

	// Read all directory entries (1 per running container)
	entries, err := os.ReadDir(dataPath)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		containerDir := filepath.Join(dataPath, entry.Name())

		// read the file container-info.txt to get the pod-name, pod-namespace, container-id fields
		info, exists := utils.ReadPropertiesFile(filepath.Join(containerDir, "container-info.txt"))
		if !exists {
			continue
		}
		namespace := info["pod-namespace"]
		podName := info["pod-name"]
		containerID := info["container-id"]

		runtimeInfo := types.ContainerRuntimeInfo{}
		// read the file os.txt to get the Operating System fingerprint
		osFingerprintPath := filepath.Join(containerDir, "os.txt")
		if info, exists := utils.ReadPropertiesFile(osFingerprintPath); exists {
			runtimeInfo.Os = utils.HashString(hash, h, info["os-release-id"])
			runtimeInfo.OsVersion = utils.HashString(hash, h, info["os-release-version-id"])
		}
		// read the file runtime-kind.txt to get the Runtime Kind fingerprint
		runtimeKindPath := filepath.Join(containerDir, "runtime-kind.txt")
		if info, exists := utils.ReadPropertiesFile(runtimeKindPath); exists {
			runtimeInfo.Kind = utils.HashString(hash, h, info["runtime-kind"])
			runtimeInfo.KindVersion = utils.HashString(hash, h, info["runtime-kind-version"])
			runtimeInfo.KindImplementer = utils.HashString(hash, h, info["runtime-kind-implementer"])
		}

		// Read all other fingerprints files to fill the runtimes map
		entries, err := os.ReadDir(containerDir)
		if err != nil {
			continue
		}
		for _, file := range entries {
			if !file.IsDir() && strings.HasSuffix(file.Name(), "-fingerprints.txt") {
				log.Println("Got fingerprints file ", filepath.Join(containerDir, file.Name()))
				if info, exists := utils.ReadPropertiesFile(filepath.Join(containerDir, file.Name())); exists {
					for k, v := range info {
						log.Println("Got key=value", k, v)

						runtimeInfo.Runtimes = append(runtimeInfo.Runtimes, types.RuntimeComponent{
							Name:    utils.HashString(hash, h, k),
							Version: utils.HashString(hash, h, v),
						})
					}
				}
			}
		}

		if _, exists := payload[namespace]; !exists {
			payload[namespace] = make(types.NamespaceRuntimeInfo)
		}
		if _, exists := payload[namespace][podName]; !exists {
			payload[namespace][podName] = make(types.PodRuntimeInfo)
		}
		payload[namespace][podName][containerID] = runtimeInfo
	}

	return payload, nil
}

// buildTLSConfigFromProfile builds a tls.Config from the OpenShift TLS security profile
func buildTLSConfigFromProfile(profile *configv1.TLSSecurityProfile) (*tls.Config, error) {
	spec, err := resolveProfileSpec(profile)
	if err != nil {
		return nil, err
	}

	minVer, err := parseTLSVersion(string(spec.MinTLSVersion))
	if err != nil {
		return nil, fmt.Errorf("invalid MinTLSVersion: %w", err)
	}

	cfg := &tls.Config{MinVersion: minVer}

	// TLS 1.3 cipher suites are not configurable in Go (golang/go#29349)
	if minVer < tls.VersionTLS13 {
		suites := mapCipherSuites(spec.Ciphers)
		if len(suites) == 0 {
			return nil, fmt.Errorf("no valid cipher suites found in TLS profile")
		}
		cfg.CipherSuites = suites
	}

	return cfg, nil
}

func resolveProfileSpec(profile *configv1.TLSSecurityProfile) (*configv1.TLSProfileSpec, error) {
	if profile == nil {
		return configv1.TLSProfiles[configv1.TLSProfileIntermediateType], nil
	}

	switch profile.Type {
	case configv1.TLSProfileOldType,
		configv1.TLSProfileIntermediateType,
		configv1.TLSProfileModernType:
		return configv1.TLSProfiles[profile.Type], nil
	case configv1.TLSProfileCustomType:
		if profile.Custom == nil {
			return nil, fmt.Errorf("custom profile specified but Custom is nil")
		}
		return &profile.Custom.TLSProfileSpec, nil
	default:
		return configv1.TLSProfiles[configv1.TLSProfileIntermediateType], nil
	}
}

func parseTLSVersion(v string) (uint16, error) {
	versions := map[string]uint16{
		"VersionTLS10": tls.VersionTLS10,
		"VersionTLS11": tls.VersionTLS11,
		"VersionTLS12": tls.VersionTLS12,
		"VersionTLS13": tls.VersionTLS13,
	}
	if ver, ok := versions[v]; ok {
		return ver, nil
	}
	return 0, fmt.Errorf("unknown TLS version: %s", v)
}

// mapCipherSuites converts OpenSSL-style cipher names to Go crypto/tls constants
func mapCipherSuites(names []string) []uint16 {
	m := map[string]uint16{
		"ECDHE-RSA-AES128-GCM-SHA256":   tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		"ECDHE-ECDSA-AES128-GCM-SHA256": tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		"ECDHE-RSA-AES256-GCM-SHA384":   tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		"ECDHE-ECDSA-AES256-GCM-SHA384": tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
		"ECDHE-RSA-CHACHA20-POLY1305":   tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
		"ECDHE-ECDSA-CHACHA20-POLY1305": tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
		"ECDHE-RSA-AES128-SHA256":       tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256,
		"ECDHE-ECDSA-AES128-SHA256":     tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256,
		"ECDHE-RSA-AES128-SHA":          tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
		"ECDHE-ECDSA-AES128-SHA":        tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
		"ECDHE-RSA-AES256-SHA":          tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
		"ECDHE-ECDSA-AES256-SHA":        tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
		"AES128-GCM-SHA256":             tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
		"AES256-GCM-SHA384":             tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
		"AES128-SHA256":                 tls.TLS_RSA_WITH_AES_128_CBC_SHA256,
		"AES128-SHA":                    tls.TLS_RSA_WITH_AES_128_CBC_SHA,
		"AES256-SHA":                    tls.TLS_RSA_WITH_AES_256_CBC_SHA,
		"DES-CBC3-SHA":                  tls.TLS_RSA_WITH_3DES_EDE_CBC_SHA,
	}

	out := make([]uint16, 0, len(names))
	for _, name := range names {
		if id, ok := m[name]; ok {
			out = append(out, id)
		}
	}
	return out
}

// watchAndExitOnTLSChange watches for TLS profile changes and exits the process
func watchAndExitOnTLSChange(configClient configclientset.Interface, current *configv1.TLSSecurityProfile) {
	w, err := configClient.ConfigV1().APIServers().Watch(context.Background(), metav1.ListOptions{
		FieldSelector: "metadata.name=cluster",
	})
	if err != nil {
		log.Printf("Failed to watch APIServer for TLS changes: %v", err)
		return
	}
	defer w.Stop()

	for event := range w.ResultChan() {
		if event.Type != watch.Modified {
			continue
		}
		updated, ok := event.Object.(*configv1.APIServer)
		if !ok {
			continue
		}
		if !equality.Semantic.DeepEqual(current, updated.Spec.TLSSecurityProfile) {
			log.Println("TLS profile changed, exiting for restart")
			os.Exit(0)
		}
	}
}

func main() {
	bindAddress := flag.String("bind", "0.0.0.0", "Bind address")
	tlsCert := flag.String("tls-cert", "", "Path to TLS certificate file (PEM format)")
	tlsKey := flag.String("tls-key", "", "Path to TLS private key file (PEM format)")
	tlsCAFile := flag.String("tls-ca", "", "Path to CA certificate for extractor TLS verification")
	tlsServerName := flag.String("tls-server-name", "", "Server name for extractor TLS certificate verification")

	flag.Parse()

	if *tlsCert == "" || *tlsKey == "" || *tlsCAFile == "" || *tlsServerName == "" {
		log.Fatal("--tls-cert, --tls-key, --tls-ca, and --tls-server-name flags are required")
	}

	extractorTLSCA = *tlsCAFile
	extractorTLSServerName = *tlsServerName

	// Fetch TLS profile from API Server
	restConfig, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("Failed to get in-cluster config: %v", err)
	}
	configClient, err := configclientset.NewForConfig(restConfig)
	if err != nil {
		log.Fatalf("Failed to create config client: %v", err)
	}

	apiserver, err := configClient.ConfigV1().APIServers().Get(context.Background(), "cluster", metav1.GetOptions{})
	if err != nil {
		log.Fatalf("Failed to get APIServer config: %v", err)
	}

	tlsConfig, err := buildTLSConfigFromProfile(apiserver.Spec.TLSSecurityProfile)
	if err != nil {
		log.Fatalf("Failed to build TLS config: %v", err)
	}

	http.HandleFunc("/gather_runtime_info", gatherRuntimeInfo)

	address := *bindAddress + ":8000"
	server := &http.Server{
		Addr:      address,
		TLSConfig: tlsConfig,
	}

	// Watch for TLS profile changes in background
	go watchAndExitOnTLSChange(configClient, apiserver.Spec.TLSSecurityProfile)

	log.Printf("Starting exporter HTTPS server at %s\n", address)
	if err := server.ListenAndServeTLS(*tlsCert, *tlsKey); err != nil {
		log.Fatal(err)
	}
}
