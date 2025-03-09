# hive-operator

This repository contains a proof of concept for the hive operator. The
operator is responsible for loading the eBPF programs into the kernel
and instruct them to log accesses to files. More specifically, the
eBPF program will log calls to the kernel funcion `inode_persmission`
which is called every time a process checks the permissions of an
inode. This happens, for example, when opening or reading a file.
The distributed nature of kubernetes clusters makes the design of the
application more challenging, this repository explains the design
considerations and proposes a possible solution.

Information in the eBPF program can be found
[here](https://github.com/San7o/hive-bpf/), this repository focuses
on the kubernetes operator.

# Status

This project is divided into multiple developement sessions:
- [DONE](https://github.com/San7o/hive-bpf/) setting up the eBPF
  program and loader outside kubernetes
- [DONE] setting up the operator and reading the container's PIDs
- [TODO] loading the eBPF program in kubernetes and tracing the files
- [TODO] logging the captured information
- [TODO] update the operator based on changes in the cluster
- [TODO] heavy testing
  - kernel versions
  - architectures
  - container runtimes
  - kubernetes versions

# Usage

To use an operator, you need a kubernetes cluster. To create one with
[kind](https://github.com/kubernetes-sigs/kind),
you can use the script `registry-cluster.sh` which will create
a cluster with one control node and one worker node. Additionally,
It sets up a local docker registry to push the operator's image
during developement. Note that you may need to use `sudo` to run the
following commands based on your permissions.

```bash
./registry-cluster.sh
```

You can delete the cluster with `delete-cluster.sh` when you do not
want It anymore.

To generate the RBAC policies, run:
```bash
make generate
```

To create the CRD manifests, run:
```bash
make manifests
```

Note that you need to run the previous two commands only if you
actually changed something that needs to be regenerated.

To build everything inside a docker container, run:
```bash
./docker-build.sh
```

This creates a docker image with the operator inside, and pushes It
int the local registry. Finally, you can deploy everything with:
```bash
make deploy
```

## Testing

Currently, the operator will log the PIDs of the container's processes
of kubernete's Pods.

Inside the operator container, you can test the application by
reading the inode of a file with the `stat` command located in a
filesystem of a pod, by reading in `/host/proc/<PID>/root/<FILE>`,
with one of the logged PIDs.

# Writeup

## Getting the inode and device id

The first challenge is getting the pair inode + device id of the file
we want to trace. We need both to identify a real ""physical"" file
because the inode is an identifier in a superblock, and different
partitions have different superblocks. Therefore, It is possible to have
different files in different partitions with the same inode. Moreover,
containers often have some ways of sandboxing between the host
filesystem and the containerized filesystem so this problem can
occur more often. 

On a UNIX system, getting this information can be done by running the
`stat(1)` command or calling `stat(2)` with the path of the file, this
is the easy part. The hard part is to get the path itself.

Considering that the operator lives inside a kubelet in a kubernetes
cluster, we cannot assume that It can access all the filasystem's of
all the pods. In fact, It can read at most the filesystems in the
kubelet where the operator lives. We cannot even assume that kubelets
live in the same machine. This means that there must be an operator on
each kubelet that accesses the filesystems of the pods present in Its
kubelet. 

The next problem is how do we decide which pod we need to access and
how. We will answer those two questions separately.

### How we decide which pod to access?

First, we need to define how do we filter containers. This can be
done throught custom kubernetes configurations, utually using namespaces
and labels. This is something kubernetes understand so we need to
query kubernetes itself for the pods that match those filters. This
is quite easy from a kubernetes client in go, we just need to call the
right function:

```go
podList := &corev1.PodList{}
if err := c.List(ctx, podList); err != nil {
	return err
}
```

But we need permissions to do so:

```go
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=pods/status,verbs=get
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=deployments/status,verbs=get
```

### How do we access the filesystem of a pod?
 
This depends on the container runtime used. On containerd, we
can access `/proc/<CONTAINER_PID>/root` to read the filesystem,
quite handy, except the fact that we need to share `/proc` between
the Pods, kubelet and operator:

```yaml
spec:
  [...]
  replicas: 1
  template:
    spec:
		[...]
        volumeMounts:
          - name: proc
            mountPath: /host/proc
            readOnly: true
	volumes:
		[...]
	- name: proc
	  hostPath:
		path: /proc
		type: Directory
```
  
To get the `PID` we need to rely on the API of the container
runtime. Thankfully, the
[OCI specification](https://opencontainers.org/) specifies that
the Pid can be accessed from the [State](https://github.com/opencontainers/runtime-spec/blob/2f2d37e8216b8019067a63c28f711482820025c6/specs-go/state.go#L51)
type, so we can ask our runtime API, if It complies to OCI, and
containerd does.
However, this means that we need to access the kubelet's containerd
daemon from the operator container:
```yaml
spec:
  [...]
  replicas: 1
  template:
    spec:
		[...]
		volumeMounts:
          - name: containerd-sock
            mountPath: /run/containerd/containerd.sock
            readOnly: false
	volumes:
	- name: containerd-sock
	  hostPath:
		path: /run/containerd/containerd.sock
		type: Socket
```

And we can create a go client:
```go
import (
	// Currently only containerd is supported
	containerd "github.com/containerd/containerd"
)

...

containerdAddress := "/run/containerd/containerd.sock"
namespace := "k8s.io"
opt := containerd.WithDefaultNamespace(namespace)
containerdClient, err := containerd.New(containerdAddress, opt)
if err != nil {
	return err
}
defer containerdClient.Close()
```

Note that the container engine and kubernetes Pods are two separate
things, but containers are identified by the same `containerID`,
so we can get the `contianerID`s from the filtered Pods and pass them
to the container runtime to get the PIDs.

A simpler solution would be to execute shell commands on the Pod,
but this would not work with distroless-containers.
