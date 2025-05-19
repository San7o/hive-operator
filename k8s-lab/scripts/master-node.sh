#!/bin/sh

source ./config.sh

if ! [ -f $FEDORA_IMAGE ]; then
    echo "ERROR: Fedora image not found in current directory."
    exit 1
fi
if ! [ -f $CLOUD_INIT_DIR/master/meta-data ]; then
    echo "ERROR: master/meta-data file not found in current directory."
    exit 1
fi
if ! [ -f $CLOUD_INIT_DIR/master/user-data ]; then
    echo "ERROR: master/user-data file not found in current directory."
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

genisoimage -output $VM_DIR/seed-master.iso \
            -volid cidata \
            -joliet \
            -rock $CLOUD_INIT_DIR/master/user-data $CLOUD_INIT_DIR/master/meta-data

if ! [ -f $VM_DIR/kube-master.qcow2 ]; then
    cp $FEDORA_IMAGE $VM_DIR/kube-master.qcow2
fi

sudo qemu-system-x86_64 \
  -enable-kvm \
  -m 2048 \
  -smp 2 \
  -cpu host \
  -drive file=$VM_DIR/kube-master.qcow2,format=qcow2 \
  -drive file=$VM_DIR/seed-master.iso,format=raw,index=1,if=virtio \
  -netdev tap,id=net0,ifname=tap0,script=no,downscript=no \
  -device virtio-net-pci,netdev=net0,mac=52:54:00:12:34:56 \
  -nographic
