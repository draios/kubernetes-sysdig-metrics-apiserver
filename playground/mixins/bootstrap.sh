#!/usr/bin/env bash

set -o pipefail
set -o errexit

export DEBIAN_FRONTEND=noninteractive

readonly ipaddr="$1"
readonly host="$2"
readonly sysdig_access_key="$3"

# if [ -z "${sysdig_access_key}" ]; then
#   # TODO: deploy agent automagically?
# fi

if ! grep vagrant /home/vagrant/.ssh/authorized_keys > /dev/null; then
  exit
fi

echo "=> Disabling swap"
swapoff -a
sed -i '/ swap / s/^/#/' /etc/fstab

echo "=> Setting noop scheduler"
echo noop > /sys/block/sda/queue/scheduler

echo "=> Disabling IPv6"
echo "net.ipv6.conf.all.disable_ipv6 = 1" >> /etc/sysctl.conf
echo "net.ipv6.conf.default.disable_ipv6 = 1" >> /etc/sysctl.conf
echo "net.ipv6.conf.lo.disable_ipv6 = 1" >> /etc/sysctl.conf
sysctl -p

echo "=> Pass bridged IPv4 traffic to iptables' chains"
modprobe br_netfilter
echo "br_netfilter" >> /etc/modules
echo "net.bridge.bridge-nf-call-iptables = 1" >> /etc/sysctl.conf
sysctl -p

echo "=> Setting up repositories"
apt-get update && apt-get install -y apt-transport-https curl ca-certificates
curl -fsSL https://apt.kubernetes.io/doc/apt-key.gpg | apt-key add -
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | apt-key add -
echo "deb http://apt.kubernetes.io/ kubernetes-xenial main" > /etc/apt/sources.list.d/kubernetes.list
echo "deb [arch=amd64] https://download.docker.com/linux/ubuntu xenial stable" > /etc/apt/sources.list.d/docker.list

echo "=> Installing dependencies"
apt-get update && apt-get install -y \
  docker-ce=$(apt-cache madison docker-ce | grep 18.03 | head -1 | awk '{print $3}') \
  kubeadm kubelet kubectl \
  ifupdown-extra

# <TODO>
# Add "-H tcp://172.17.8.101:2376" to Docker's systemd unit configuration so
# the developer can access to the Docker daemon from the host. This is useful
# during the development workflow, e.g. `skaffold dev`.
# </TODO>

echo "=> Configuring kubelet arguments"
CGROUP_DRIVER=$(sudo docker info | grep "Cgroup Driver" | awk '{print $3}')
sed -i "s|KUBELET_KUBECONFIG_ARGS=|KUBELET_KUBECONFIG_ARGS=--cgroup-driver=$CGROUP_DRIVER |g" /etc/systemd/system/kubelet.service.d/10-kubeadm.conf
systemctl daemon-reload
systemctl stop kubelet && systemctl start kubelet

echo "=> Installing kernel headers (needed by Sysdig Agent)"
apt-get -qq -y install linux-headers-$(uname -r)

# Ensure that `hostname -i` returns a routable IP address (i.e. one on the
# second network interface, not the first one). By default, it doesn’t do
# this and kubelet ends-up using first non-loopback network interface, which
# is usually NATed. Workaround: override /etc/hosts.
# See https://kubernetes.io/docs/setup/independent/troubleshooting-kubeadm/.
echo "=> Ensuring that 'hostname -i' returns a routable IP address"
echo "${ipaddr} ${host}" > /etc/hosts

# This is a workaround for Vagrant setups where we have two network interfaces
# and the Kubernetes components are not reachable on the default route. We add
# a custom route so the Kubernetes cluster addresses go via the appropiate
# adapter.
# Related links:
# - https://kubernetes.io/docs/setup/independent/install-kubeadm/#check-network-adapters
# - https://github.com/kubernetes/kubeadm/issues/102
echo "=> Ensuring that the Kubernetes cluster addresses are reachable"
echo "10.96.0.0 255.240.0.0 172.17.8.100 eth1" > /etc/network/routes
ip route add 10.96.0.0/12 dev eth1
