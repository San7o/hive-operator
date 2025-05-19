# Test cluster

This directory contains useful scripts to create a local test
kubernetes cluster. It uses qemu with KVM to create a master VM and
some worker VMs connected to the host and to eachother via a tap
interface through a bridge. Kubernetes is installed using
[k3s](https://k3s.io/) on Fedora Cloud images.

The following commands available to build the cluster:

- qemu-system-x86_64
- genisoimage
- ip
- tmux

First, you need to setup the network interfaces to make the VMs talk
to eachother and to the internet:

```bash
make setup
```

You need to download the fedora cloud image from their
[website](https://fedoraproject.org/cloud/) and place It in the
`images` directory. You need to modify the entry `FEDORA_IMAGE` on
[config.sh](./config.sh) with the name of the image.

Finally you can deploy the cluster with:

```bash
make
```

To undo the changes run:

```bash
make clean
make network-reset
```
