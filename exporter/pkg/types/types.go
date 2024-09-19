package types

type NodeRuntimeInfo map[string]NamespaceRuntimeInfo
type NamespaceRuntimeInfo map[string]PodRuntimeInfo
type PodRuntimeInfo map[string]ContainerRuntimeInfo

// ContainerRuntimeInfo represents workload info returned by the insights-runtime-extractor component.
type ContainerRuntimeInfo struct {
	// Hash of the identifier of the Operating System (based on /etc/os-release ID)
	Os string `json:"os,omitempty"`
	// Hash of the version identifier of the Operating System (based on /etc/os-release VERSION_ID)
	OsVersion string `json:"osVersion,omitempty"`
	// Identifier of the kind of runtime
	Kind string `json:"kind,omitempty"`
	// Version of the kind of runtime
	KindVersion string `json:"kindVersion,omitempty"`
	// Entity that provides the runtime-kind implementation
	KindImplementer string `json:"kindImplementer,omitempty"`
	// Runtimes components
	Runtimes []RuntimeComponent `json:"runtimes,omitempty"`
}

type RuntimeComponent struct {
	// Name of a runtime used to run the application in the container
	Name string `json:"name,omitempty"`
	// The version of this runtime
	Version string `json:"version,omitempty"`
}
