#! /usr/bin/env bash
set -eu

source hack/tools/install.sh
source ./hack/ci/handy.sh
KUBECONFIG=$(pwd)/hack/ci/secrets/gardener-kubeconfig.yaml
export KUBECONFIG

# get the latest kubernetes version marked as supported in the cloudprofile
SHOOT_VERSION=$(kubectl get cloudprofile regiocloud -o yaml | yq '.spec.kubernetes.versions[] | select(.classification == "supported") | .version' | sort -V | tail -n1)
export SHOOT_VERSION

yq '.metadata.name=env(SHOOT_NAME) | .spec.kubernetes.version=env(SHOOT_VERSION)' hack/ci/misc/shoot-for-seed.yaml | kubectl apply -f -

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
