#!/usr/bin/env bash

set -o pipefail
set -o errexit

[[ -n ${DEBUG} ]] && set -x

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

[[ -n "${SSH_CONFIG}" ]] && SSH_OPTIONS=("${SSH_OPTIONS[@]}" "-F" "${SSH_CONFIG}")
[[ -n "${SSH_KEYFILE}" ]] && SSH_OPTIONS=("${SSH_OPTIONS[@]}" "-i" "${SSH_KEYFILE}")

do_scp() {
    scp "${SSH_OPTIONS[@]}" "$@"
}

do_ssh() {
    ssh "${SSH_OPTIONS[@]}" "$@"
}

set_up_controller() {
    local instance="controller-node"
    do_ssh "${VM_USER}@${instance}" sudo kubeadm reset
    do_ssh "${VM_USER}@${instance}" sudo kubeadm init --config=/vagrant/mixins/kubeadm.yml
}

create_token() {
    local instance="controller-node"
    local cmd=$(do_ssh "${VM_USER}@${instance}" sudo kubeadm token create --print-join-command)
    echo ${cmd}
}

kubeconfig() {
    local instance="controller-node"
    tmpdir=$(mktemp -d)
    cd "$tmpdir"
    do_ssh "${VM_USER}@${instance}" "sudo cat /etc/kubernetes/admin.conf" > "${tmpdir}/config"
    echo ${tmpdir}
}

patch_kube_proxy() {
    kubectl -n kube-system get ds -l "k8s-app=kube-proxy" -o json \
        | jq '.items[0].spec.template.spec.containers[0].command |= .+ ["--cluster-cidr=10.32.0.0/12"]' \
        | kubectl apply -f - && kubectl -n kube-system delete pods -l "component=kube-proxy"
}

# Initialize the machine with the control plane components.
set_up_controller

# Extract kubeconfig.
kubeconfig_dir=$(kubeconfig)

# Export KUBECONFIG so we can use kubectl in this script.
export KUBECONFIG="${kubeconfig_dir}/config"

# Install Weave Net.
# patch_kube_proxy - (I'm not entirely sure if I need this, I don't think so!)
export kubever=$(kubectl version | base64 | tr -d '\n')
kubectl apply -f "https://cloud.weave.works/k8s/net?k8s-version=${kubever}"

# Join nodes.
join_cmd=$(create_token)
for i in $(seq 1 "$WORKERS"); do
    instance="worker-node-${i}"
    do_ssh "${VM_USER}@${instance}" sudo kubeadm reset
    do_ssh "${VM_USER}@${instance}" "sudo ${join_cmd}"
done

echo "We're done, thank you for waiting!"
echo "kubectl config available in ${kubeconfig_dir}/config"
echo
echo "Usage example:"
echo "  $ export KUBECONFIG=${kubeconfig_dir}/config"
echo "  $ kubectl get nodes"
echo "-------------------------------------------------------------------------"
