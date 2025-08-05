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

	corev1 "k8s.io/api/core/v1"

	kivev1alpha1 "github.com/San7o/kivebpf/api/v1alpha1"
)

const (
	// The mountpoint inside the operator's container of the node's
	// procfs
	ProcMountpoint = "/host/proc"
	// If the node is a container (for example, this happens with
	// clusters created with Kind), the actual host's procfs is assumed
	// to be mounted here
	RealHostProcMountpoint = "/host/real/proc"
)

type ContainerName = string
type ContainerID = string
type Ino = uint64
type Dev = uint64
type Pid = uint32

type ContainerData struct {
	Ino  Ino
	ID   string
	Name ContainerName
	// If true, ContainerData should be requested again later
	ShouldRequeue bool
	// False if an inode was not found, used for improved error messages
	IsFound bool
}

type Runtime interface {
	IsConnected() bool
	Connect(ctx context.Context) error
	Disconnect() error
	GetContainerData(ctx context.Context, id ContainerID, kiveTrap kivev1alpha1.KiveTrap) (ContainerData, error)
}

var (
	ContainerRuntimes map[string]Runtime = make(map[string]Runtime)
)

func init() {
	ContainerRuntimes = map[string]Runtime{
		"containerd": &Containerd{}, // Currently only containerd is supported
	}
}

func GetContainerData(ctx context.Context, containerStatus corev1.ContainerStatus, kiveTrap kivev1alpha1.KiveTrap) (ContainerData, error) {

	if !containerStatus.Ready {
		return ContainerData{ShouldRequeue: true}, nil
	}

	runtimeName, containerId, err := SplitContainerRuntimeID(containerStatus.ContainerID)
	if err != nil {
		return ContainerData{}, err
	}
	supported := IsContainerRuntimeSupported(runtimeName)
	if !supported {
		return ContainerData{}, fmt.Errorf("GetContainerData Error: Container runtime %s is not suported.", runtimeName)
	}
	runtime := ContainerRuntimes[runtimeName]

	if !runtime.IsConnected() {
		if err := runtime.Connect(ctx); err != nil {
			return ContainerData{}, fmt.Errorf("GetContainerData Error Connect: %w", err)
		}
	}

	containerData, err := runtime.GetContainerData(ctx, containerId, kiveTrap)
	if err == nil {
		containerData.ID = containerStatus.ContainerID
		containerData.Name = containerStatus.Name
		return containerData, nil
	}

	return ContainerData{}, nil
}

func CloseConnections() error {

	for _, containerRuntime := range ContainerRuntimes {

		if !containerRuntime.IsConnected() {
			continue
		}

		if err := containerRuntime.Disconnect(); err != nil {
			return fmt.Errorf("Container CloseConnections Error: %w", err)
		}
	}

	return nil
}
