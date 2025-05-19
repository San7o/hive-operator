#!/bin/sh

source ./config.sh

if ! [ -f $FEDORA_IMAGE ]; then
    echo "ERROR: Fedora image not found in current directory."
    exit 1
fi
if ! [ -f $CLOUD_INIT_DIR/worker1/meta-data ]; then
    echo "ERROR: worler1/meta-data file not found in current directory."
    exit 1
fi
if ! [ -f $CLOUD_INIT_DIR/worker1/user-data ]; then
    echo "ERROR: worker1/user-data file not found in current directory."
    exit 1
fi
if ! type genisoimage > /dev/null; then
    echo "ERROR: genisoimage command not found."
    exit 1
fi
if ! type qemu-system-x86_64 > /dev/null; then
    echo "ERROR: qemu-system-x86_64 command not found."
    exit 1
fi

genisoimage -output $VM_DIR/seed-worker1.iso \
            -volid cidata \
            -joliet \
            -rock $CLOUD_INIT_DIR/worker1/user-data $CLOUD_INIT_DIR/worker1/meta-data

if ! [ -f $VM_DIR/kube-worker1.qcow2 ]; then
    cp $FEDORA_IMAGE $VM_DIR/kube-worker1.qcow2
fi

sudo qemu-system-x86_64 \
  -enable-kvm \
  -m 2048 \
  -smp 2 \
  -cpu host \
  -drive file=$VM_DIR/kube-worker1.qcow2,format=qcow2 \
  -drive file=$VM_DIR/seed-worker1.iso,format=raw,index=1,if=virtio \
  -netdev tap,id=net0,ifname=tap1,script=no,downscript=no \
  -device virtio-net-pci,netdev=net0,mac=52:54:00:12:34:57 \
  -nographic
