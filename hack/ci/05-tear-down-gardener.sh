#! /usr/bin/env bash
set -eu

source hack/ci/handy.sh

cd gardener || exit

# Tear down the gardener environment
make gardener-extensions-down
make kind-extensions-clean
