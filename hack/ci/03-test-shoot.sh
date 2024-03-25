#! /usr/bin/env bash
set -eu

source hack/tools/install.sh
source hack/ci/handy.sh

# Create a test shoot now
kind get kubeconfig -n gardener-extensions > gardener-kind.yaml
export KUBECONFIG=gardener-kind.yaml
yq '.metadata.name=env(TEST_SHOOT_NAME) | .spec.kubernetes.version=env(TEST_SHOOT_VERSION)' hack/ci/misc/test-shoot.yaml | kubectl apply -f -

echo "Waiting for shoot creation..."
while [ ! "$(kubectl get shoot -n garden-project-1 "$TEST_SHOOT_NAME" -o jsonpath="{.status.lastOperation.state}")" == "Succeeded" ]; do
  PERCENTAGE=$(kubectl get shoot -n garden-project-1 "$TEST_SHOOT_NAME" -o jsonpath="{.status.lastOperation.progress}")
  echo "Creating shoot: $PERCENTAGE%"
  sleep 5
done
echo "Shoot creation succeeded"
