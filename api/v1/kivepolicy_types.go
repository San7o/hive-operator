/*
                    GNU GENERAL PUBLIC LICENSE
                       Version 2, June 1991

 Copyright (C) 1989, 1991 Free Software Foundation, Inc.,
 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301 USA
 Everyone is permitted to copy and distribute verbatim copies
 of this license document, but changing it is not allowed.
*/

// SPDX-License-Identifier: GPL-2.0-only

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	KivePolicyFinalizerName = "kivepolicy.kivebpf.san7o.github.io/finalizer"
)

type KivePolicySpec struct {
	// Version for KiveAlert output
	AlertVersion string `json:"alertVersion,omitempty"`
	// List of traps
	Traps []KiveTrap `json:"traps,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

type KivePolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec KivePolicySpec `json:"spec,omitempty"`
}

// +kubebuilder:object:root=true

type KivePolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KivePolicy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KivePolicy{}, &KivePolicyList{})
}
