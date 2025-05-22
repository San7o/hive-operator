# k8s-lab

This directory contains useful scripts to create a local test
kubernetes cluster. The cluster is composed of local virtual machines
running via qemu with KVM. The virtual machines are connected to the
host network and to themselves via tap interfaces and a bridge on the
host's system. The officially supported operating system is linux,
more specifically [Fedora Cloud](https://fedoraproject.org/cloud/).

Potentially, you could use any distro as a base as long as you can run
all the software required for a kubernetes cluster, mainly a container
runtime. In the future this could be extended to test specific kernel
versions or patches while maintaining a common userspace.

## Requirements

The following commands must be available on the host machine to build
the cluster:

- qemu-system-x86_64
- cloud-localds
- ip
- tmux
- awk

## Build the cluster

The following commands bootstrap a cluster with two virtual machines:
a master kubernetes node and a worker kubernetes node. To setup
everything you just need to run simple commands which are explained
below.

First, you need to setup the network interfaces to make the VMs talk
to each other and to the internet. This consists of setting up a tap
interface per virtual machine and a common bridge.

```bash
make setup
```

You need to download the fedora cloud image from their
[website](https://fedoraproject.org/cloud/) and place It in the
`images` directory. You also need to modify the entry `FEDORA_IMAGE`
on [config.sh](./config.sh) with the name of the downloaded image.

You need to generate the cloud-init files. Those are used to easily
configure the virtual machines with a bunch of files and the user
"fedora" with password "fedora".

```
make generate
```

Finally you can deploy the cluster with:

```bash
make
```

To undo what you just did, run:

```bash
make clean # removes the images
make network-reset # removes the tap and bridge interfaces
```

Have fun!
