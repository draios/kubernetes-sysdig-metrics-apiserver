#!/usr/bin/env bash

# Workaround until https://github.com/GoogleCloudPlatform/skaffold/issues/226 is solved.

set -o errexit
set -o pipefail
set -o nounset
# set -o xtrace

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BIN="${DIR}/bin"

# Build with cache
CGO_ENABLED=0 GOOS=linux go install -ldflags="-w -s" -v github.com/sevein/k8s-sysdig-adapter/cmd/adapter

# Copy binary in Docker's context
[ -d ${BIN} ] || mkdir ${BIN}
cp $GOPATH/bin/adapter ${BIN}/adapter

# Run skaffold
skaffold run

# Ugly but it works for now!
sleep 5
SERVER_POD=$(kubectl -n custom-metrics get pod -o jsonpath='{.items[0].metadata.name}')
kubectl -n custom-metrics logs -f ${SERVER_POD}
