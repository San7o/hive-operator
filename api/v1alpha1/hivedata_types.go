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

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// HiveDataSpec defines the desired state of HiveData
type HiveDataSpec struct {
	// The path of the file
	PathName string `json:"path-name,omitempty"`
	// The name of the pod where this file lives
	PodName string `json:"pod-name,omitempty"`
	// The namespace of the pod where the file lives
	PodNamespace string `json:"pod-namespace,omitempty"`
	// The inode number of the file
	InodeNo uint32 `json:"inode-no,omitempty"`
	// The device id of the file. Currently unsupported
	DevID uint64 `json:"dev-id,omitempty"`
	// A string to uniquely identify a running kernel
	KernelID string `json:"kernel-id,omitempty"`
}

// HiveDataStatus defines the observed state of HiveData
//type HiveDataStatus struct {
// Either "created", "removed" or "updated"
//Operation string `json:"operation,omitempty"`
//}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// HiveData is the Schema for the hivedata API
type HiveData struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec HiveDataSpec `json:"spec,omitempty"`
	//Status HiveDataStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// HiveDataList contains a list of HiveData
type HiveDataList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HiveData `json:"items"`
}

func init() {
	SchemeBuilder.Register(&HiveData{}, &HiveDataList{})
}
