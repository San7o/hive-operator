/*
                    GNU GENERAL PUBLIC LICENSE
                       Version 2, June 1991

 Copyright (C) 1989, 1991 Free Software Foundation, Inc.,
 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301 USA
 Everyone is permitted to copy and distribute verbatim copies
 of this license document, but changing it is not allowed.
*/

// SPDX-License-Identifier: GPL-2.0-only

package v1alpha1

// Information about the container
type ContainerMetadata struct {
	// Container id
	Id string `json:"id,omitempty"`
	// Container name
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

// Information about the node
type NodeMetadata struct {
	// Name of the node
	Name string `json:"name,omitempty"`
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
	Cwd string `json:"cwd,omitempty"`
}

// Additional information
type HiveAlertMetadata struct {
	// File path
	Path string `json:"path,omitempty"`
	// Inode number of the file
	Inode uint64 `json:"inode,omitempty"`
	// Unix access permission mask
	Mask int32 `json:"mask,omitempty"`
	// ID of the kernel where the alert was triggered
	KernelID string `json:"kernel_id,omitempty"`
	// Callback URI
	Callback string `json:"callback,omitempty"`
}

// File access alert
type HiveAlert struct {
	// Alert creation time
	Timestamp string `json:"timestamp,omitempty"` // RFC 3339
	// The policy that triggered the alert
	HivePolicyName string `json:"hive_policy_name,omitempty"`
	// Additional information
	Metadata HiveAlertMetadata `json:"metadata,omitempty"`
	// Information about the pod where the file lives
	Pod PodMetadata `json:"pod,omitempty"`
	// Information about the node
	Node NodeMetadata `json:"node,omitempty"`
	// Information about the process that accessed the file
	Process ProcessMetadata `json:"process,omitempty"`
}
