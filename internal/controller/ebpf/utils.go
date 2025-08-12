/*
                    GNU GENERAL PUBLIC LICENSE
                       Version 2, June 1991

 Copyright (C) 1989, 1991 Free Software Foundation, Inc.,
 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301 USA
 Everyone is permitted to copy and distribute verbatim copies
 of this license document, but changing it is not allowed.
*/

// SPDX-License-Identifier: GPL-2.0-only

package ebpf

import (
	"fmt"
	"strings"

	"github.com/cilium/ebpf"
)

/*
 *  Add an entry to the map
 */
func AddInode(mapKey BpfMapKey) error {

	one := uint8(1)
	err := Objs.TracedInodes.Update(mapKey, one, ebpf.UpdateAny)
	if err != nil {
		return fmt.Errorf("AddInode Error: %w", err)
	}

	return nil
}

/*
 *  Remove an entry from the map
 */
func RemoveInode(mapKey BpfMapKey) error {

	zero := uint8(0)
	err := Objs.TracedInodes.Update(mapKey, zero, ebpf.UpdateAny)
	if err != nil {
		return fmt.Errorf("RemoveInode Error: %w", err)
	}

	return nil
}

func int8ArrayToString(arr [16]int8) string {
	b := make([]byte, 0, len(arr))
	for _, c := range arr {
		if c == 0 {
			break
		}
		b = append(b, byte(c))
	}
	return string(b)
}

func parseCmdline(cmdline string) (binary string, args string) {

	parts := strings.Split(cmdline, "\x00")

	cleaned := make([]string, 0, len(parts))
	for _, part := range parts {
		if part != "" {
			cleaned = append(cleaned, part)
		}
	}

	if len(cleaned) == 0 {
		return "", ""
	}

	binary = cleaned[0]
	if len(cleaned) > 1 {
		args = strings.Join(cleaned[1:], " ")
	} else {
		args = ""
	}

	return binary, args
}
