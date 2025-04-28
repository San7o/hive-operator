# Developement

This documents contains useful information to build the operator
yourself. It explains how to create a local test cluster and build
both the eBPF program and the kubernetes operator. To test the build,
please read the [USAGE](./USAGE.md) document.

## Setup the cluster

To use an operator, you need a kubernetes cluster. This repository
provides the script `registry-cluster.sh` which will create a cluster
using [kind](https://github.com/kubernetes-sigs/kind) with one control
node and one worker node. Additionally, It sets up a local docker
registry to push the operator's image during developement.

Run the following command to create the cluster (needs to be run only
once):

```bash
make create-cluster-local
```

You can delete the cluster with `delete-cluster.sh` or with
`make delete-cluster` when you do not want It anymore.

## Generate files

The operator uses generators to create various config files such as
RBAC policies, CRD manifests and the eBPF program. Those need to be
regenerated every time they are updates.

To generate the RBAC policies, run:

```bash
make generate
```

To create the CRD manifests, run:

```bash
make manifests
```

To generate the eBPF program, first you need to have the following
dependencies in your system:

- Linux kernel version 5.7 or later, with ebpf support enabled
- LLVM 11 or later (clang and llvm-strip)
- libbpf headers
- Linux kernel headers
- a recent go compiler

Once you have the dependencies, run:

```bash
make generate-ebpf
```

## Build the docker contianer

Note that you may need sudo priviledges for the following commands
based on your system.

To build everything inside a docker container, first **make sure that
generated files are updated**, then build the docker image with:

```bash
make docker-build-local
```

You can push it to a test local docker repository if you generated
the cluster with `registry-cluster.sh`:

```bash
make docker-push-local
```

To do both of the above in a single command, run:

```bash
make docker-local
```

## Deploy the operator

Finally, you can deploy the operator to the test cluster with:

```bash
make deploy
```

You can now proceed by reading the [USAGE](./USAGE.md) document which
will explain how to use the operator.

## Useful commands

When building a new docker image, you want the hive pods to update to
the new version. The pods are already configured to fetch the latest
version of the image in the local test docker repository, to do so you
just need to kill them. You can use the following command:

```bash
make kill-pods-local
```

This will also remove all the HiveData resources so that you start
with a clean configuration, as if you just applied the HivePolicies.
