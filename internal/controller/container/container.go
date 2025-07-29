package container

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"

	hivev1alpha1 "github.com/San7o/hive-operator/api/v1alpha1"
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
	GetContainerData(ctx context.Context, id ContainerID, hivePolicy hivev1alpha1.HivePolicy) (ContainerData, error)
}

var (
	ContainerRuntimes map[string]Runtime = make(map[string]Runtime)
)

func init() {
	ContainerRuntimes = map[string]Runtime{
		"containerd": &Containerd{}, // Currently only containerd is supported
	}
}

func GetContainerData(ctx context.Context, containerStatus corev1.ContainerStatus, hivePolicy hivev1alpha1.HivePolicy) (ContainerData, error) {

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

	containerData, err := runtime.GetContainerData(ctx, containerId, hivePolicy)
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
