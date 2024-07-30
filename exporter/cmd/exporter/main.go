package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"exporter/pkg/types"
	"exporter/pkg/utils"
)

const (
	EXTRACTOR_ADDRESS string = "127.0.0.1:3000"
)

// gatherRuntimeInfo will trigger a new extraction of runtime info
// and reply with a JSON payload
func gatherRuntimeInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method is not supported.", http.StatusNotFound)
		return
	}

	hashParam := r.URL.Query().Get("hash")
	hash := hashParam == "" || hashParam == "true"

	dataPath, err := triggerRuntimeInfoExtraction()
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
	log.Println("Reading runtime info data from :", dataPath)

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
	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}

func triggerRuntimeInfoExtraction() (string, error) {
	conn, err := net.Dial("tcp", EXTRACTOR_ADDRESS)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	log.Println("Trigger new runtime extraction")
	// write to TCP connection to trigger a runtime extraction
	fmt.Fprintf(conn, "")

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

func main() {
	http.HandleFunc("/gather-runtime-info", gatherRuntimeInfo)

	log.Println("Starting exporter HTTP server at port 8000")
	if err := http.ListenAndServe(":8000", nil); err != nil {
		log.Fatal(err)
	}
}
