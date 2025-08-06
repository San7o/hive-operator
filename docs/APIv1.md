# APIv1

Here is the stable `v1` api available since version `1.0.0` of the
operator.

## KivePolicy

```go
type KivePolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec KivePolicySpec `json:"spec,omitempty"`
}

type KivePolicySpec struct {
	// Version for KiveAlert output
	AlertVersion string `json:"alertVersion,omitempty"`
	// List of traps
	Traps []KiveTrap `json:"traps,omitempty"`
}

type KiveTrap struct {
	// Specifies which path to monitor
	Path string `json:"path,omitempty"`
	// (optional) Whether to create the file or not if It was not found
	Create bool `json:"create,omitempty"`
	// (optional) The permissions of the file to be created if create is set to true
	Mode uint32 `json:"mode,omitempty"`
	// (optional) Send an HTTP POST request to this endpoint
	Callback string `json:"callback,omitempty"`
	// Match any of the following items (logical OR), at least one must be present
	MatchAny []KiveTrapMatch `json:"matchAny,omitempty"`
}

// Match all the following optional fields (logical AND)
type KiveTrapMatch struct {
	// Filter pods by name
	PodName string `json:"pod,omitempty"`
	// Filter container by name, can be a regex with syntax described at
	// https://golang.org/s/re2syntax
	ContainerName string `json:"containerName,omitempty"`
	// Filter pods by namespace
	Namespace string `json:"namespace,omitempty"`
	// Filter pods by IP
	IP string `json:"ip,omitempty"`
	// Filter pods by label
	MatchLabels map[string]string `json:"matchLabels,omitempty"`
}
```

## KiveAlert

```go
// File access alert
type KiveAlert struct {
	// KiveAlert version
	AlertVersion string `json:"kive-alert-version"`
	// The policy that triggered the alert
	PolicyName string `json:"kive-policy-name,omitempty"`
	// Alert creation time
	Timestamp string `json:"timestamp,omitempty"` // RFC 3339
	// Additional information
	Metadata KiveAlertMetadata `json:"metadata,omitempty"`
	// Information about the pod where the file lives
	Pod PodMetadata `json:"pod,omitempty"`
	// Information about the node
	Node NodeMetadata `json:"node,omitempty"`
	// Information about the process that accessed the file
	Process ProcessMetadata `json:"process,omitempty"`
}

// Additional information
type KiveAlertMetadata struct {
	// File path
	Path string `json:"path,omitempty"`
	// Inode number of the file
	Inode uint64 `json:"inode,omitempty"`
	// Unix access permission mask
	Mask int32 `json:"mask,omitempty"`
	// ID of the kernel where the alert was triggered
	KernelID string `json:"kernel-id,omitempty"`
	// Callback URI
	Callback string `json:"callback,omitempty"`
}

// Information related to the process that accessed the file
type ProcessMetadata struct {
	// Process ID
	Pid int32 `json:"pid"`
	// Thread group ID
	Tgid uint32 `json:"tgid"`
	// User ID
	Uid uint32 `json:"uid"`
	// Group ID
	Gid uint32 `json:"gid"`
	// Process binary
	Binary string `json:"binary,omitempty"`
	// Current Working Directory
	Cwd string `json:"cwd"`
    // Arguments to the Binary
	Arguments string `json:"arguments"`
}

// Information about the node
type NodeMetadata struct {
	// Name of the node
	Name string `json:"name,omitempty"`
}

// Information about the pod where the file lives
type PodMetadata struct {
	// Pod name
	Name string `json:"name,omitempty"`
	// Pod namespace
	Namespace string `json:"namespace,omitempty"`
	// Pod ip
	Ip string `json:"ip,omitempty"`
	// Information about the container
	Container ContainerMetadata `json:"container,omitempty"`
}

// Information about the container
type ContainerMetadata struct {
	// Container id
	Id string `json:"id,omitempty"`
	// Container name
	Name string `json:"name,omitempty"`
}
```
