package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type nodeRuntimeInfo map[string]namespaceRuntimeInfo
type namespaceRuntimeInfo map[string]podRuntimeInfo
type podRuntimeInfo map[string]containerRuntimeInfo

// containerRuntimeInfo represents workload info returned by the insights-runtime-extractor component.
type containerRuntimeInfo struct {
	OSReleaseID            string             `json:"os-release-id,omitempty"`
	OSReleaseVersionID     string             `json:"os-release-version-id,omitempty"`
	RuntimeKind            string             `json:"runtime-kind,omitempty"`
	RuntimeKindVersion     string             `json:"runtime-kind-version,omitempty"`
	RuntimeKindImplementer string             `json:"runtime-kind-implementer,omitempty"`
	Runtimes               []RuntimeComponent `json:"runtimes,omitempty"`
}

type RuntimeComponent struct {
	// Name of a runtime used to run the application in the container
	Name string `json:"name,omitempty"`
	// The version of this runtime
	Version string `json:"version,omitempty"`
}

// gatherRuntimeInfo will trigger a new extraction of runtime info
// and reply with a JSON payload
func gatherRuntimeInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method is not supported.", http.StatusNotFound)
		return
	}

	// connect to the exporter server on the loopback address
	conn, err := net.Dial("tcp", "127.0.0.1:3000")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	// trigger a runtime extraction by opening a connection to the exporter server
	fmt.Fprintf(conn, "")

	dataPath, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	dataPath = strings.TrimSpace(dataPath)
	fmt.Println("Reading runtime info data from :", dataPath)

	payload, err := collect_workload_payload(dataPath)
	fmt.Println("Payload:", payload)
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

	// remove the dataPath directory
}

func collect_workload_payload(dataPath string) (nodeRuntimeInfo, error) {
	// create a map with the key being "${pod-namespace}-${pod-name}-${container-id}"
	payload := make(nodeRuntimeInfo)

	fmt.Printf("Reading data from |%s|\n", dataPath)
	// Read all directory entries (1 per running container)
	entries, err := os.ReadDir(dataPath)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			containerDir := filepath.Join(dataPath, entry.Name())

			// read the file container-info.txt to get the pod-name, pod-namespace, container-id fields
			info, exists := readPropertiesFile(filepath.Join(containerDir, "container-info.txt"))
			if !exists {
				continue
			}
			namespace := info["pod-namespace"]
			podName := info["pod-name"]
			containerID := info["container-id"]

			fmt.Println("Reading info for container:", namespace, podName, containerID)

			runtimeInfo := containerRuntimeInfo{}
			osFingerprintPath := filepath.Join(containerDir, "os.txt")
			if info, exists := readPropertiesFile(osFingerprintPath); exists {
				runtimeInfo.OSReleaseID = info["os-release-id"]
				runtimeInfo.OSReleaseVersionID = info["os-release-version-id"]
			}
			runtimeKindPath := filepath.Join(containerDir, "runtime-kind.txt")
			if info, exists := readPropertiesFile(runtimeKindPath); exists {
				runtimeInfo.RuntimeKind = info["runtime-kind"]
				runtimeInfo.RuntimeKindVersion = info["runtime-kind-version"]
				runtimeInfo.RuntimeKindImplementer = info["runtime-kind-implementer"]
			}

			// Read the directory
			entries, err := os.ReadDir(containerDir)
			if err != nil {
				continue
			}

			var runtimes []RuntimeComponent

			for _, file := range entries {
				if !file.IsDir() && strings.HasSuffix(file.Name(), "-fingerprints.txt") {
					if info, exists := readPropertiesFile(filepath.Join(containerDir, file.Name())); exists {
						for k, v := range info {
							runtimes = append(runtimes, RuntimeComponent{
								Name:    k,
								Version: v,
							})
						}
					}
				}
			}
			runtimeInfo.Runtimes = runtimes

			if _, exists := payload[namespace]; !exists {
				fmt.Println("create entry for namespace", namespace)
				payload[namespace] = make(namespaceRuntimeInfo)
			}
			if _, exists := payload[namespace][podName]; !exists {
				fmt.Println("create entry for namespace pod", namespace, podName)
				payload[namespace][podName] = make(podRuntimeInfo)
			}
			payload[namespace][podName][containerID] = runtimeInfo
		}
	}

	return payload, nil
}

func readPropertiesFile(filePath string) (map[string]string, bool) {
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		// File does not exist
		return nil, false
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	properties := make(map[string]string)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") || strings.TrimSpace(line) == "" {
			continue
		}

		// Split the line into key and value
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			properties[key] = value
		}
	}

	// Check for scanner errors
	if err := scanner.Err(); err != nil {
		return nil, false
	}

	return properties, true
}

func readFfileExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		// File exists
		return true
	}
	if os.IsNotExist(err) {
		// File does not exist
		return false
	}
	// Some other error occurred
	return false
}

func main() {
	// Set up HTTP handlers for specific paths.
	http.HandleFunc("/gather-runtime-info", gatherRuntimeInfo)

	// Start the HTTP server on port 3000.
	fmt.Println("Starting server at port 8000")
	if err := http.ListenAndServe(":8000", nil); err != nil {
		fmt.Println(err)
	}
}
