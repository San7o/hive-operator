#!/bin/bash
# SPDX-License-Identifier: GPL-2.0-only

set -e

echo "Tearing down network..."

for i in 0 1; do
  tap="tap$i"
  sudo ip link set "$tap" down 2>/dev/null || true
  sudo ip tuntap del "$tap" mode tap 2>/dev/null || true
done

sudo ip link set br0 down 2>/dev/null || true
sudo ip link delete br0 type bridge 2>/dev/null || true

#sudo iptables -t nat -D POSTROUTING -s 192.168.50.0/24 ! -o br0 -j MASQUERADE 2>/dev/null || true

echo "Teardown complete."
