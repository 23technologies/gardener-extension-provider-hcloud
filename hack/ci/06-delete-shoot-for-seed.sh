#! /usr/bin/env bash
set -eu

source hack/ci/handy.sh

# Delete the shoot on okeanos.dev
export KUBECONFIG=hack/ci/secrets/gardener-kubeconfig.yaml
kubectl annotate shoot "$SHOOT_NAME" confirmation.gardener.cloud/deletion=true --overwrite=true || echo "Annotating shoot failed"
kubectl delete shoot "$SHOOT_NAME" --wait=false || echo "Deleting shoot failed"
