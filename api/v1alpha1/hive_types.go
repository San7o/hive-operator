/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
