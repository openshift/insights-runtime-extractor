package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"insights-runtime-exporter/pkg/types"
	"insights-runtime-exporter/pkg/utils"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
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
	fmt.Println("Reading runtime info data from :", dataPath)

	payload, err := collectWorkloadPayload(hash, dataPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Println("Payload:", payload)

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

	fmt.Println("Trigger new runtime extraction")
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

		fmt.Println("Reading info for container:", namespace, podName, containerID)

		runtimeInfo := types.ContainerRuntimeInfo{}
		// read the file os.txt to get the Operating System fingerprint
		osFingerprintPath := filepath.Join(containerDir, "os.txt")
		if info, exists := utils.ReadPropertiesFile(osFingerprintPath); exists {
			runtimeInfo.OSReleaseID = utils.HashString(hash, h, info["os-release-id"])
			runtimeInfo.OSReleaseVersionID = utils.HashString(hash, h, info["os-release-version-id"])
		}
		// read the file runtime-kind.txt to get the Runtime Kind fingerprint
		runtimeKindPath := filepath.Join(containerDir, "runtime-kind.txt")
		if info, exists := utils.ReadPropertiesFile(runtimeKindPath); exists {
			runtimeInfo.RuntimeKind = utils.HashString(hash, h, info["runtime-kind"])
			runtimeInfo.RuntimeKindVersion = utils.HashString(hash, h, info["runtime-kind-version"])
			runtimeInfo.RuntimeKindImplementer = utils.HashString(hash, h, info["runtime-kind-implementer"])
		}

		if _, exists := payload[namespace]; !exists {
			payload[namespace] = make(types.NamespaceRuntimeInfo)
		}
		if _, exists := payload[namespace][podName]; !exists {
			payload[namespace][podName] = make(types.PodRuntimeInfo)
		}
		payload[namespace][podName][containerID] = runtimeInfo

		// Read all other fingerprints files to fill the runtimes map
		var runtimes []types.RuntimeComponent
		entries, err := os.ReadDir(containerDir)
		if err != nil {
			continue
		}
		for _, file := range entries {
			if !file.IsDir() && strings.HasSuffix(file.Name(), "-fingerprints.txt") {
				if info, exists := utils.ReadPropertiesFile(filepath.Join(containerDir, file.Name())); exists {
					for k, v := range info {
						runtimes = append(runtimes, types.RuntimeComponent{
							Name:    utils.HashString(hash, h, k),
							Version: utils.HashString(hash, h, v),
						})
					}
				}
			}
		}
		runtimeInfo.Runtimes = runtimes
	}

	return payload, nil
}

func main() {
	http.HandleFunc("/gather-runtime-info", gatherRuntimeInfo)

	fmt.Println("Starting exporter HTTP server at port 8000")
	if err := http.ListenAndServe(":8000", nil); err != nil {
		fmt.Println(err)
	}
}
