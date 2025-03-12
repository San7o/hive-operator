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

func GetInodeDevID(pid Pid, path string) (Ino, Dev, error) {
	pidStr := strconv.FormatUint(uint64(pid), 10)
	target := procMountpoint + separator + pidStr +
		separator + "root" + separator + path
	var stat syscall.Stat_t
	err := syscall.Stat(target, &stat)
	if err != nil {
		return uint64(0), uint64(0), err
	}

	return stat.Ino, stat.Dev, nil
}
