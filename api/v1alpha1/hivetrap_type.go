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

type HiveTrap struct {
	// Specifies which path to check
	Path string `json:"path,omitempty"`
	// Whether to create the file or not if It cannot be found
	Create bool `json:"create,omitempty"`
	// Send an HTTP POST request to this endpoint if specified
	Callback string `json:"callback,omitempty"`
	// The content of the file if It was created. This field
	// is used only if Create is set to true
	Mode uint32 `json:"mode,omitempty"`
	// Filters the pods inside this namespace
	Match HiveTrapMatch `json:"match,omitempty"`
}

type HiveTrapMatch struct {
	// Filter pod by name
	PodName string `json:"pod,omitempty"`
	// Filter container by name, can be a regex with syntax described at
	// https://golang.org/s/re2syntax
	ContainerName string `json:"container-name,omitempty"`
	// Filter pods per namespace
	Namespace string `json:"namespace,omitempty"`
	// Filter pods by IP
	IP string `json:"ip,omitempty"`
	// Filter pods per label
	MatchLabels map[string]string `json:"matchLabels,omitempty"`
}
