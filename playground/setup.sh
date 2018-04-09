#!/usr/bin/env bash

set -o pipefail
set -o errexit

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# The default setup uses three virtual machines.
# One controller and N=2 workers. Tweak the number of workers below.
export WORKERS=${WORKERS:-2}
export SSH_CONFIG="${DIR}/../.vagrant/ssh_config"
export SSH_KEYFILE="${HOME}/.vagrant.d/insecure_private_key"
export VM_USER="vagrant"
export VAGRANT_CWD="${DIR}"
export SYSDIG_ACCESS_KEY="${SYSDIG_ACCESS_KEY}"

if [[ $1 == "clean" ]]; then
  vagrant destroy -f
else
  vagrant up
  vagrant ssh-config > "${SSH_CONFIG}"
  "${DIR}/mixins/kubeadm.sh"
fi
