#!/bin/sh
# SPDX-License-Identifier: GPL-2.0-only

# Configure the test cluster here

export FEDORA_IMAGE=images/Fedora-Cloud-Base-Generic-42-1.1.x86_64.qcow2
export SSH_PORT_MASTER=2222
export SSH_PORT_WORKER1=2223
export VM_DIR=vm
export CLOUD_INIT_DIR=cloud-init
export SCRIPTS_DIR=scripts
export SESSION_NAME="k8s-lab"
