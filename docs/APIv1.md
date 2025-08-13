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
    // (optional) Additional information for this trap
	Metadata string `json:"metadata,omitempty"`
	// Match any of the following items (logical OR), at least one must be present
	MatchAny []KiveTrapMatch `json:"matchAny,omitempty"`
}

// Match all the following optional fields (logical AND)
type KiveTrapMatch struct {
    // Filter pods by name
    PodName string `json:"pod,omitempty"`
    // Filter container by name.
    //  - if this field is prepended by "regex:", the rest of the string
    //    will represent a regular expression matched with go regexp
    //    library (https://golang.org/s/re2syntax)
    //  - if the field is prepended by "glob:", then this is a
    //    filesystem-style regex, as described in go filepath.Match
    //    library (https://pkg.go.dev/path/filepath#Match)
    //  - otherwise, the name of the container will be compared exactly
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
	PolicyName string `json:"kive-policy-name"`
	// Alert creation time
	Timestamp string `json:"timestamp"` // RFC 3339
	// Additional information
	Metadata KiveAlertMetadata `json:"metadata"`
    // User specified metadata (from KivePolicy)
	CustomMetadata map[string]string `json:"custom-metadata"`
	// Information about the pod where the file lives
	Pod PodMetadata `json:"pod"`
	// Information about the node
	Node NodeMetadata `json:"node"`
	// Information about the process that accessed the file
	Process ProcessMetadata `json:"process"`
}

// Additional information
type KiveAlertMetadata struct {
	// File path
	Path string `json:"path"`
	// Inode number of the file
	Inode uint64 `json:"inode"`
	// Unix access permission mask
	Mask int32 `json:"mask"`
	// ID of the kernel where the alert was triggered
	KernelID string `json:"kernel-id"`
	// Callback URI
	Callback string `json:"callback"`
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
	Binary string `json:"binary"`
	// Current Working Directory
	Cwd string `json:"cwd"`
    // Arguments to the Binary
	Arguments string `json:"arguments"`
}

// Information about the node
type NodeMetadata struct {
	// Name of the node
	Name string `json:"name"`
}

// Information about the pod where the file lives
type PodMetadata struct {
	// Pod name
	Name string `json:"name"`
	// Pod namespace
	Namespace string `json:"namespace"`
	// Pod ip
	Ip string `json:"ip"`
	// Information about the container
	Container ContainerMetadata `json:"container"`
}

// Information about the container
type ContainerMetadata struct {
	// Container id
	Id string `json:"id"`
	// Container name
	Name string `json:"name"`
}
```
