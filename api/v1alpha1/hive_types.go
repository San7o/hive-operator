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

type HivePolicy struct {
	// Specifies which path to check
	Path string `json:"path,omitempty"`
	// Whether to create the file or not if It cannot be found
	Create bool `json:"create,omitempty"`
	// The content of the file if It was created. This field
	// is used only if Create is set to true
	Content string `json:"content,omitempty"`
	// Filters the pods inside this namespace
	Match HivePolicyMatch `json:"match,omitempty"`
}

type HivePolicyMatch struct {
	// Filter pod by name
	Pod string `json:"pod,omitempty"`
	// Filter pods per namespace
	Namespace string `json:"namespace,omitempty"`
	// Filter pods per label
	Label string `json:"label,omitempty"`
}

// HiveStatus defines the observed state of Hive
type HiveStatus struct {
	// Either "create" "update" "delete"
	Operation string `json:"operation,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Hive is the Schema for the hives API
type Hive struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HivePolicy `json:"spec,omitempty"`
	Status HiveStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// HiveList contains a list of Hive
type HiveList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Hive `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Hive{}, &HiveList{})
}
