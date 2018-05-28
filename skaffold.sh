#!/usr/bin/env bash

# Deployment script for developers.
#
# It assumes that:
# - kubectl's current context is pointing to your cluster.
# - docker's client is pointing to the daemon in the worker node.
#
# I use this script together with the `playground`.
#
# Temporary solution until I can use skaffold which I tried but it seems pretty
# unstable and it does not support incremental builds yet - I'm really liking
# its ideas though so I'm hoping to switch as soon as possible.

set -o errexit
set -o pipefail
set -o nounset

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BIN="${DIR}/bin"

IMAGE="sevein/k8s-sysdig-adapter"
AGENT_KEY=${AGENT_KEY:?Need to set AGENT_KEY}
SDC_TOKEN=${SDC_TOKEN:?Need to set SDC_TOKEN}

# This should be doing incremental builds if you're using Go v1.10.
env CGO_ENABLED=0 GOOS=linux go build -v -ldflags="-w -s" -o ${BIN}/adapter -v github.com/draios/kubernetes-sysdig-metrics-apiserver/cmd/adapter

# Build image
docker build -f Dockerfile.skaffold -t ${IMAGE} .

# Tag it using its checksum
tag=$(docker inspect --format='{{index .Id}}' sevein/k8s-sysdig-adapter | sed -e "s/^sha256://" | cut -c 10)
docker tag ${IMAGE} "${IMAGE}:${tag}"

# Prepare configuration and install it
export AGENT_KEY=$(echo -n ${AGENT_KEY} | base64)
export SDC_TOKEN=$(echo -n ${SDC_TOKEN} | base64)
export IMAGE="${IMAGE}:${tag}"
envsubst < ${DIR}/playground/mixins/deployment.yml.tmpl | kubectl apply -f -

# Watch logs
kubectl logs -n custom-metrics deployment/custom-metrics-apiserver -f
