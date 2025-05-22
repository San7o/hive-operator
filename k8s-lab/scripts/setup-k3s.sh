#!/bin/sh
# SPDX-License-Identifier: GPL-2.0-only

# This script currently does not work but I think It would be possible
# to make k3s work with some time. I had problems with certificates
# and permissions.

K3S_FILE=/tmp/k3s.sh

print_usage()
{
    echo "Usage: <master|worker>"
    return
}

download_k3s()
{
    curl -o $K3S_FILE -sfL https://get.k3s.io
}

if [ $# -le 0 ]; then
    print_usage
    exit 1
fi

if [ $1 = "master" ]; then
    echo "Setting up master node..."
    download_k3s
    K3S_TOKEN="k8s-lab" sh $K3S_FILE
    exit 0
fi
if [ $1 = "worker" ]; then
    if [ $# -le 2 ] || [ $# -ge 3 ]; then
        echo "Usage: worker K3S_SERVER_IP"
        exit 1
    fi

    echo "Setting up worker node..."
    download_k3s
    K3S_URL=https://$2:6441 K3S_TOKEN="k8s-lab" sh $K3S_FILE    
    exit 0
fi

echo "Error: Command not recognized"
print_usage
exit 1
