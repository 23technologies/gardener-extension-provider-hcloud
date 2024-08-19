#! /usr/bin/env bash
set -eu

source hack/tools/install.sh
source hack/ci/handy.sh

# Create a test shoot now
kind get kubeconfig -n gardener-extensions > gardener-kind.yaml
export KUBECONFIG=gardener-kind.yaml
yq '.metadata.name=env(TEST_SHOOT_NAME) | .spec.kubernetes.version=env(TEST_SHOOT_VERSION)' hack/ci/misc/test-shoot.yaml | kubectl apply -f -

MAX_TRIES=300
TRY=0
WAIT=5

echo "Waiting for shoot creation..."
while [ ! "$(kubectl get shoot -n garden-project-1 "$TEST_SHOOT_NAME" -o jsonpath="{.status.lastOperation.state}")" == "Succeeded" ]; do
  (( TRY+=1 ))

  if [[ $TRY -gt $MAX_TRIES ]]; then
      echo "Shoot creation timed out after $((WAIT * MAX_TRIES)) seconds"
      exit 1;
  fi

  PERCENTAGE=$(kubectl get shoot -n garden-project-1 "$TEST_SHOOT_NAME" -o jsonpath="{.status.lastOperation.progress}")
  echo "Creating shoot: $PERCENTAGE%"
  sleep $WAIT
done
echo "Shoot creation succeeded"
