package types

type NodeRuntimeInfo map[string]NamespaceRuntimeInfo
type NamespaceRuntimeInfo map[string]PodRuntimeInfo
type PodRuntimeInfo map[string]ContainerRuntimeInfo

// containerRuntimeInfo represents workload info returned by the insights-runtime-extractor component.
type ContainerRuntimeInfo struct {
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
