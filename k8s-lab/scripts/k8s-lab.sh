#!/bin/sh
# SPDX-License-Identifier: GPL-2.0-only

KUBERNETES_VERSION_MAJOR=1
KUBERNETES_VERSION_MINOR=33
KUBERNETES_VERSION_PATCH=0

CONTROL_PLANE_IP="192.168.50.10"
CONTROL_PLANE_PORT="6443"
PROXY_PORT=8080
export KUBECONFIG=/etc/kubernetes/admin.conf
FLANNEL_LINK=https://github.com/flannel-io/flannel/releases/latest/download/kube-flannel.yml
SCRIPT_NAME=k8s-lab

set -e

print_usage()
{
    echo "Usage: $SCRIPT_NAME <master|join|delete|token|deploy|undeploy>"
    echo ""
    echo "  master    initialize the master node"
    echo "  join      join a cluster a worker node"
    echo "  delete    delete the master node"
    echo "  token     generate the command from the master for a worker to join"
    echo "  deploy    deploy a sample nginx pod"
    echo "  undeploy  undeploy the sample nginx pod"
}

if [ "$1" = "help" ] || [ "$1" = "--help" ] || [ "1" = "-h" ]; then
    print_usage
    exit 0
elif [ "$1" = "join" ]; then
    if [ $# -le 3 ]; then
        echo "Usage: $SCRIPT_NAME join IP:PORT TOKEN HASH"
        exit 1
    fi
elif [ "$1" = "delete" ]; then
    sudo kubeadm reset -f
    sudo rm -rf /etc/cni/net.d2 || true
    exit 0
elif [ "$1" = "token" ]; then
    TOKEN=$(sudo kubeadm token list | awk 'NR==2 {print $1}')
    if [ "$TOKEN" = "" ]; then
        TOKEN=$(sudo kubeadm token generate)
        sudo kubeadm token create $TOKEN
    fi
    CA_HASH=$(openssl x509 -pubkey -in /etc/kubernetes/pki/ca.crt | openssl rsa -pubin -outform der 2>/dev/null | sha256sum | awk '{print $1}')
    if [ -z "$CA_HASH" ]; then
        echo "Failed to compute CA cert hash!"
        exit 1  
    fi

    echo "Use the following command to join a node to the cluster:"
    echo ""
    echo "sudo $SCRIPT_NAME join ${CONTROL_PLANE_IP}:${CONTROL_PLANE_PORT} ${TOKEN} sha256:${CA_HASH}"
    exit 0
elif [ "$1" = "deploy" ]; then
    sudo kubectl create deployment nginx --image=nginx
    exit 0
elif [ "$1" = "undeploy" ]; then
    sudo kubectl delete deployment nginx
    exit 0
elif [ "$1" != "master" ]; then
    print_usage
    exit 1
fi

# Maybe add another command to refresh the kubernetes configs
# using sudo kubeadm init phase kubeconfig all
# This may be useful after a reboot

echo "Swapoff"
sudo swapoff -a
sudo sed -i '/ swap / s/^/#/' /etc/fstab

echo "Disabling SELinux"
sudo setenforce 0

echo "Downloading dependencies"
sudo dnf install -y kubernetes$KUBERNETES_VERSION_MAJOR.$KUBERNETES_VERSION_MINOR-kubeadm vim containerd docker git

echo "Setting up network"
sudo modprobe br_netfilter
sudo modprobe overlay
echo '1' | sudo tee /proc/sys/net/ipv4/ip_forward
echo "br_netfilter" | sudo tee -a /etc/modules-load.d/custom-modules.conf
echo "overlay" | sudo tee -a /etc/modules-load.d/custom-modules.conf
cat <<EOF | tee /etc/sysctl.d/k8s.conf
net.bridge.bridge-nf-call-iptables  = 1
net.bridge.bridge-nf-call-ip6tables = 1
net.ipv4.ip_forward                 = 1
EOF
sudo sysctl --system

echo "Enabling containerd.service"
sudo systemctl enable containerd
sudo mkdir -p /etc/containerd
sudo containerd config default | tee /etc/containerd/config.toml
echo "Starting containerd.service"
sudo systemctl start containerd

echo "Enabling kubelet.service"
sudo systemctl enable kubelet
echo "Starting kubelet.service"
sudo systemctl start kubelet

echo "Setting KUBECONF env variable"
echo KUBECONFIG=$KUBECONFIG | sudo tee -a /etc/environment

echo "Starting kubeadm"
sudo rm -rf /etc/cni/net.d /var/lib/etcd /var/lib/kubelet /etc/kubernetes

if [ "$1" = "join" ]; then
    sudo kubeadm join $2 --token $3 --discovery-token-ca-cert-hash $4 --discovery-token-unsafe-skip-ca-verification
else
    sudo kubeadm init --pod-network-cidr=10.244.0.0/16 --ignore-preflight-errors=all --apiserver-advertise-address=$CONTROL_PLANE_IP
    
    echo "Starting proxy"
    coproc kubectl proxy --port=$PROXY_PORT

    echo "Applying flannel"
    kubectl apply -f $FLANNEL_LINK
fi

echo "Copying cni binaries"
# Thank you fedora
sudo mkdir /opt/cni
sudo mkdir /opt/cni/bin
sudo cp /usr/libexec/cni/* /opt/cni/bin/

cat <<EOF | tee -a /var/lib/kubelet/config.yaml
evictionHard:
  imagefs.available: 1%
  memory.available: 100Mi
  nodefs.available: 1%
  nodefs.inodesFree: 1%
EOF

sudo systemctl restart containerd
sudo systemctl restart kubelet

echo "setup-kubeadm done"
