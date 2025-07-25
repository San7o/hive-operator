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
	// The inode number of the file
	InodeNo uint64 `json:"inode-no,omitempty"`
	// A string to uniquely identify a running kernel
	KernelID string `json:"kernel-id,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

type HiveData struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec HiveDataSpec `json:"spec,omitempty"`
}

// +kubebuilder:object:root=true

type HiveDataList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HiveData `json:"items"`
}

func init() {
	SchemeBuilder.Register(&HiveData{}, &HiveDataList{})
}
