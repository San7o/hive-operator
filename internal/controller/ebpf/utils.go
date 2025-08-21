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
	"strconv"
	"strings"
	"syscall"

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

	err := Objs.TracedInodes.Delete(mapKey)
	if err != nil {
		return fmt.Errorf("RemoveInode Error: %w", err)
	}

	return nil
}

func int8ArrayToString(arr []int8) string {
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

/*
 *  The function check_permission was hanged in linux 5.12, hence
 *  we need to load a different program if the version is newer or
 *  older than that.
 */
func isKernelOld() (bool, error) {

	var uname syscall.Utsname
	if err := syscall.Uname(&uname); err != nil {
		return false, fmt.Errorf("isKernelOld Error Syscall Uname: %w", err)
	}

	release := int8ArrayToString(uname.Release[:])
	major, err := strconv.Atoi(strings.Split(release, ".")[0])
	if err != nil {
		return false, fmt.Errorf("isKernelOld Error Conversion to integer of major kernel version: %w", err)
	}
	minor, err := strconv.Atoi(strings.Split(release, ".")[1])
	if err != nil {
		return false, fmt.Errorf("isKernelOld Error Conversion to integer of minor kernel version: %w", err)
	}

	return major < 5 || (major == 5 && minor <= 11), nil
}
