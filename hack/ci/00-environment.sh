#! /usr/bin/env bash
set -eu

source hack/tools/install.sh

export SHOOT_HASH=$(openssl rand -hex 2)
export SHOOT_NAME=ci-seed-$SHOOT_HASH
export TEST_SHOOT_NAME=test-$SHOOT_HASH

cat << EOF > hack/ci/handy.sh
export AZURE_DNS_CLIENT_ID=$AZURE_DNS_CLIENT_ID
export AZURE_DNS_CLIENT_SECRET=$AZURE_DNS_CLIENT_SECRET
export AZURE_DNS_SUBSCRIPTION_ID=$AZURE_DNS_SUBSCRIPTION_ID
export AZURE_DNS_TENANT_ID=$AZURE_DNS_TENANT_ID
export HCLOUD_TOKEN=$HCLOUD_TOKEN

export SHOOT_NAME=$SHOOT_NAME
export TEST_SHOOT_NAME=$TEST_SHOOT_NAME
export PATH=$(pwd)/hack/tools/bin/:$PATH
EOF


if [[ ! -d gardener ]]; then
		git clone https://github.com/gardener/gardener.git
fi
cd gardener || exit
git fetch --all
git checkout "$(git tag -l 'v1.*' | sort --version-sort | tail -1)"
git checkout v1.84.0

# Waiting only for 5 minutes may be too short. Wait for 10 minutes instead
sed -i 's/elapsed_time -gt 300/elapsed_time -gt 600/' example/provider-extensions/registry-seed/deploy-registry.sh
