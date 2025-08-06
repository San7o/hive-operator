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
