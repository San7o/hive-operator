/*
                    GNU GENERAL PUBLIC LICENSE
                       Version 2, June 1991

 Copyright (C) 1989, 1991 Free Software Foundation, Inc.,
 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301 USA
 Everyone is permitted to copy and distribute verbatim copies
 of this license document, but changing it is not allowed.
*/

// SPDX-License-Identifier: GPL-2.0-only

package container

import (
	"fmt"
	"strings"
)

func SplitContainerRuntimeID(input ContainerName) (ContainerName, ContainerID, error) {
	// input is of the form "<type>://<container_id>".
	// For example, the type could be "containerd"
	split := strings.SplitN(input, "://", 2)

	if len(split) != 2 {
		return "", "", fmt.Errorf("Error parsing containerID")
	}

	var id ContainerID = split[1]
	return split[0], id, nil
}

func IsContainerRuntimeSupported(runtime string) bool {
	for containerName, _ := range ContainerRuntimes {
		if runtime == containerName {
			return true
		}
	}
	return false
}
