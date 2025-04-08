#!/bin/sh

ARCH=$(uname -m)
KVERSION=$(uname -r)
KMAJOR=$(echo $KVERSION | cut -d'.' -f1)
KMINOR=$(echo $KVERSION | cut -d'.' -f2)
FILENAME=$ARCH/vmlinux_${ARCH}_${KMAJOR}_${KMINOR}.h

if [ ! -d $ARCH ]; then
    mkdir $ARCH
fi

bpftool btf dump file /sys/kernel/btf/vmlinux format c > $FILENAME

echo File $FILENAME generated
