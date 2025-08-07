/*
                    GNU GENERAL PUBLIC LICENSE
                       Version 2, June 1991

 Copyright (C) 1989, 1991 Free Software Foundation, Inc.,
 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301 USA
 Everyone is permitted to copy and distribute verbatim copies
 of this license document, but changing it is not allowed.
*/

// SPDX-License-Identifier: GPL-2.0-only

package v2alpha1

var (
	SupportedKiveAlertVersions = []string{"v1", "v2alpha1"}
)

// Information about the container
type ContainerMetadata struct {
	// Container id
	Id string `json:"id"`
	// Container name
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

// Information about the node
type NodeMetadata struct {
	// Name of the node
	Name string `json:"name"`
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
	// Information about the pod where the file lives
	Pod PodMetadata `json:"pod"`
	// Information about the node
	Node NodeMetadata `json:"node"`
	// Information about the process that accessed the file
	Process ProcessMetadata `json:"process"`
}
