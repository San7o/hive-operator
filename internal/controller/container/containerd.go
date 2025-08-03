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
	"context"
	"fmt"
	"strconv"
	"syscall"

	containerd "github.com/containerd/containerd"
	containerdCio "github.com/containerd/containerd/cio"

	hivev1alpha1 "github.com/San7o/hive-operator/api/v1alpha1"
)

const (
	containerdAddress = "/run/containerd/containerd.sock"
	separator         = "/"
	namespace         = "k8s.io"
)

type Containerd struct {
	Client      *containerd.Client
	isConnected bool
}

func (self *Containerd) Connect(ctx context.Context) error {

	var err error
	if self.Client != nil {
		serving, err := self.Client.IsServing(ctx)
		if err != nil || !serving {
			return fmt.Errorf("Containerd Connect Error IsServing(): %w", err)
		}
	} else {
		opt := containerd.WithDefaultNamespace(namespace)
		self.Client, err = containerd.New(containerdAddress, opt)
		if err != nil {
			return fmt.Errorf("Containerd Connect Error New connection: %w", err)
		}
	}

	self.isConnected = true

	return nil
}

func (self *Containerd) Disconnect() error {

	if err := self.Client.Close(); err != nil {
		return fmt.Errorf("Containerd Disconnect Error Close: %w", err)
	}

	self.isConnected = false

	return nil
}

func (self *Containerd) IsConnected() bool {
	return self.isConnected
}

func (self *Containerd) GetContainerData(ctx context.Context, id string, hiveTrap hivev1alpha1.HiveTrap) (ContainerData, error) {

	attach := containerdCio.NewAttach()

	containers, err := self.Client.Containers(ctx)
	if err != nil {
		return ContainerData{}, err
	}

	for _, container := range containers {
		if container.ID() == id {
			task, err := container.Task(ctx, attach)
			if err != nil {
				return ContainerData{}, err
			}

			inode, err := getInode(task.Pid(),
				hiveTrap.Path, hiveTrap.Create, hiveTrap.Mode)
			if err != nil {
				return ContainerData{}, err
			}

			return ContainerData{Ino: inode, IsFound: true}, nil
		}
	}

	return ContainerData{}, fmt.Errorf("Containerd GetContainerData Inode not found")
}

/*
 *  Get the inode of a file, given the path and the pid of the
 *  container where the file lives. Creates the file with mode
 *  permissions if create is set to true.
 */
func getInode(pid Pid, path string, create bool, mode uint32) (Ino, error) {
	pidStr := strconv.FormatUint(uint64(pid), 10)
	target := ProcMountpoint + separator + pidStr +
		separator + "root" + separator + path
	var stat syscall.Stat_t

	if create {
		fd, err := syscall.Creat(target, mode)
		if err != nil {
			return uint64(0), err
		}
		syscall.Close(fd)
	}

	err := syscall.Stat(target, &stat)
	if err != nil {
		return uint64(0), err
	}

	return stat.Ino, nil
}
