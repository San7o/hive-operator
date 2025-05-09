/*
                    GNU GENERAL PUBLIC LICENSE
                       Version 2, June 1991

 Copyright (C) 1989, 1991 Free Software Foundation, Inc.,
 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301 USA
 Everyone is permitted to copy and distribute verbatim copies
 of this license document, but changing it is not allowed.
*/

// SPDX-License-Identifier: GPL-2.0-only

package controller

import (
	"errors"
	"strconv"
	"strings"
	"syscall"
)

type ContainerRuntime = string
type ContainerID = string
type Ino = uint64
type Dev = uint64
type Pid = uint32

const (
	procMountpoint                  = "/host/proc"
	separator                       = "/"
	Containerd     ContainerRuntime = "containerd"
)

// TODO: Add support to more container runtimes
var SupportedContainerRuntimes []ContainerRuntime = []ContainerRuntime{Containerd}

func SplitContainerRuntimeID(input string) (ContainerRuntime, ContainerID, error) {
	// input is of the form "<type>://<container_id>".
	// For example, the type could be "containerd"
	split := strings.SplitN(input, "://", 2)

	if len(split) != 2 {
		return "", "", errors.New("Error parsing containerID")
	}

	var id ContainerID = split[1]
	return split[0], id, nil
}

func IsContainerRuntimeSupported(runtime ContainerRuntime) bool {
	var supported = false
	for _, it := range SupportedContainerRuntimes {
		if runtime == it {
			supported = true
			break
		}
	}
	if supported {
		return true
	}
	return false
}

func GetInodeDevID(pid Pid, path string, create bool, mode uint32) (Ino, Dev, error) {
	pidStr := strconv.FormatUint(uint64(pid), 10)
	target := procMountpoint + separator + pidStr +
		separator + "root" + separator + path
	var stat syscall.Stat_t

	if create {
		fd, err := syscall.Creat(target, mode)
		if err != nil {
			return uint64(0), uint64(0), err
		}
		syscall.Close(fd)
	}

	err := syscall.Stat(target, &stat)
	if err != nil {
		return uint64(0), uint64(0), err
	}

	return stat.Ino, stat.Dev, nil
}
