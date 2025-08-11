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
 *  Update element at index with value in the TracedInodes map.
 */
func UpdateTracedInodes(index uint32, value uint64) error {

	err := Objs.TracedInodes.Update(index, value, ebpf.UpdateAny)
	if err != nil {
		return fmt.Errorf("UpdateTracedInodes Error: %w", err)
	}

	return nil
}

/*
 * Fills a map with zeroes from index to MapMaxEntries.
 */
func ResetTracedInodes(index uint32) error {

	for ; index < MapMaxEntries; index++ {
		err := UpdateTracedInodes(index, uint64(0))
		if err != nil {
			return fmt.Errorf("ResetMap Error: %w", err)
		}
	}

	return nil
}

/*
 *  Remove an inode from the map
 */
func RemoveInode(inode uint64) error {

	var value uint64
	for i := 0; i < MapMaxEntries; i++ {
		key := uint32(i)

		err := Objs.TracedInodes.Lookup(&key, &value)
		if err != nil {
			continue
		}

		if value == inode {
			var zero uint64 = 0
			if err := Objs.TracedInodes.Update(&key, &zero, ebpf.UpdateAny); err != nil {
				return fmt.Errorf("RemoveInode: failed to clear key %d: %v", key, err)
			}
		}
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
