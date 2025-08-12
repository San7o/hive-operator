/*
                    GNU GENERAL PUBLIC LICENSE
                       Version 2, June 1991

 Copyright (C) 1989, 1991 Free Software Foundation, Inc.,
 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301 USA
 Everyone is permitted to copy and distribute verbatim copies
 of this license document, but changing it is not allowed.
*/

// SPDX-License-Identifier: GPL-2.0-only

package main

import (
	"fmt"
	"syscall"

	container "github.com/San7o/kivebpf/internal/controller/container"
)

func GetInodeDev(Target string) (uint64, uint32, error) {

	fd, err := syscall.Open(Target, syscall.O_RDONLY, 444)
	if err != nil {
		return 0, 0, fmt.Errorf("GetInode Error Open: %w", err)
	}
	defer syscall.Close(fd)

	var stat syscall.Stat_t
	err = syscall.Fstat(fd, &stat)
	if err != nil {
		return 0, 0, fmt.Errorf("GetInode Error Fstat: %w", err)
	}

	fmt.Printf("Inode: %d\n", stat.Ino)
	fmt.Printf("Dev: %d\n", container.UserDevToKernelDev(stat.Dev))

	return stat.Ino, container.UserDevToKernelDev(stat.Dev), nil
}
