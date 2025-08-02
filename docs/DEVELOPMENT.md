# Developement

This documents contains useful information to build the operator
yourself. It explains how to create a local test cluster and build
both the eBPF program and the kubernetes operator. To test the build,
please read the [USAGE](./USAGE.md) document.

If you want to develop locally, the following sections are
helpful. However, you could skip them and jump directly to the
deployment if you just want to download the images from docker hub
instead of building them.

## Setup a local cluster

To use an operator, you need a kubernetes cluster. This repository
provides the script `registry-cluster.sh` which will create a local
cluster using [kind](https://github.com/kubernetes-sigs/kind) with one
control node and one worker node. Additionally, It sets up a local
docker registry to push the operator's image during development.

If you are using Kind, or any other method where the node runs inside
a container, you need to mount `/proc` inside the node's container in
`/host/real/proc`. The script above already does this. This is needed
because the eBPF program runs in the host's kernel and the operator
needs to access the host's procfs to generate the `HiveAlert`. If you
do not do this, the operator will gracefully report a message and some
fields in the alert will remain empty (namely, `cwd`). If you are
using virtual machines or hypervisors, this is not needed.

Run the following command to create the cluster (needs to be run only
once):

```bash
make create-cluster-local
```

You can delete the cluster with `delete-cluster.sh` or with
`make delete-cluster-local` when you do not want It anymore.

## Dev Environments

For convenience, this project uses different environments to manage
building and deploying. By default, there are two different
environments: `local` and `remote`.

You can easily add a custom environment by creating a file called
`.env-<ENV-NAME>` where `<ENV-NAME>` is a name of your choice. This
file will be included in the Makefile before running any command, so
you can change the variables used by the Makefile from the env file
without changing the Makefile.

For example, the `IMG` variable tells the Makefile where to push
images and tells the operator where to pull them. You can add an entry
to your custom environment .env-custom like so:

```
IMG=registry/my-bautiful-name:latest
```

To select which environment to use, append `ENV=<ENV-NAME>` after your
make commands, for example:

```
make deploy ENV=custom
```

The default environment is `local`, in this case you can omit the `ENV`
from the make command to use the local environment. You can use
environments on all make commands.

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

On Ubuntu, you can run the following command to install the required
dependencies:

```bash
apt-get install make clang llvm libbpf-dev golang linux-headers-$(uname -r)
```

Once you have the dependencies, run:

```bash
make generate-ebpf
```

## Build the docker contianer

Note that you may need sudo privileges for the following commands
depending on your permissions.

To build everything inside a docker container, first **make sure that
generated files are updated** (section above), then build the docker
image with:

```bash
make docker-build
```

You can push it to a test local docker repository if you generated
the cluster with `registry-cluster.sh`:

```bash
make docker-push
```

To do both of the above in a single command, run:

```bash
make docker
```

## Deploy the operator

Finally, you can deploy the operator to the cluster with:

```bash
make deploy
```

You can now proceed by reading the [USAGE](./USAGE.md) document which
will explain how to use the operator.

## Testing

To run the end to end test, first make sure that you have a cluster
running with the operator deployed, and that there is not `HivePolicy`
present. Then, simply run:

```
make test
```

## Useful commands

When building a new docker image, you want the hive pods to update to
the new version. The pods are already configured to fetch the latest
version of the image in the local test docker repository, to do so you
just need to kill them. You can use the following command:

```bash
make kill-pods
```

This will also remove all the` HiveData` resources so that you start
with a clean configuration, as if you just applied the `HivePolicies`.

To completely remove the operator from the cluster, run:

```bash
make undeploy
```
