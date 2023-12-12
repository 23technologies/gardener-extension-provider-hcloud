#! /usr/bin/env bash
set -eu

source ./hack/ci/handy.sh
export KUBECONFIG=$(pwd)/hack/ci/secrets/gardener-kubeconfig.yaml


yq '.metadata.name=env(SHOOT_NAME)' hack/ci/misc/shoot-for-seed.yaml | kubectl apply -f -

echo "Waiting for shoot creation..."
while [ ! "$(kubectl get shoot "$SHOOT_NAME" -o jsonpath="{.status.lastOperation.state}")" == "Succeeded" ]; do
  PERCENTAGE=$(kubectl get shoot "$SHOOT_NAME" -o jsonpath="{.status.lastOperation.progress}")
  echo "Creating shoot: $PERCENTAGE%"
  sleep 5
done
echo "Shoot creation succeeded"

# This will get us the kubeconfig for our seed cluster
kubectl create \
    -f <(printf '{"spec":{"expirationSeconds":360000}}') \
    --raw /apis/core.gardener.cloud/v1beta1/namespaces/garden-23ke-ci/shoots/"${SHOOT_NAME}"/adminkubeconfig | \
    jq -r ".status.kubeconfig" | \
    base64 -d > gardener/example/provider-extensions/seed/kubeconfig
