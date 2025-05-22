#!/bin/bash
# SPDX-License-Identifier: GPL-2.0-only

set -e

echo "Creating bridge br0..."

sudo ip link set br0 down 2>/dev/null || true
sudo ip link delete br0 type bridge 2>/dev/null || true

sudo ip link add name br0 type bridge
sudo ip addr add 192.168.50.1/24 dev br0
sudo ip link set br0 up

echo 1 | sudo tee /proc/sys/net/ipv4/ip_forward
sudo iptables -t nat -A POSTROUTING -s 192.168.50.0/24 ! -o br0 -j MASQUERADE

for i in 0 1; do
  tap="tap$i"

  echo "Setting up $tap..."

  sudo ip link set "$tap" down 2>/dev/null || true
  sudo ip tuntap del "$tap" mode tap 2>/dev/null || true

  sudo ip tuntap add dev "$tap" mode tap
  sudo ip link set "$tap" up
  sudo ip link set "$tap" master br0
done

echo "Bridge and tap devices ready."
