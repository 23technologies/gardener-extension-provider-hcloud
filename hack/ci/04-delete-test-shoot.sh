#! /usr/bin/env bash
set -eu

source hack/tools/install.sh
source hack/ci/handy.sh

# Create a test shoot now
kind get kubeconfig -n gardener-extensions > gardener-kind.yaml
export KUBECONFIG=gardener-kind.yaml

# And delete the test-shoot again
kubectl annotate shoot -n garden-project-1 "$TEST_SHOOT_NAME" confirmation.gardener.cloud/deletion=true --overwrite=true
kubectl delete shoot -n garden-project-1 "$TEST_SHOOT_NAME" --wait=true
