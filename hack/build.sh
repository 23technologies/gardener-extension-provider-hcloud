#!/usr/bin/env bash

# Copyright (c) 2021 SAP SE or an SAP affiliate company. All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -e

# For the build step concourse will set the following environment variables:
# SOURCE_PATH - path to component repository root directory.
# BINARY_PATH - path to an existing (empty) directory to place build results into.
if [[ $(uname) == 'Darwin' ]]; then
  READLINK_BIN="greadlink"
else
  READLINK_BIN="readlink"
fi

if [[ -z "${SOURCE_PATH}" ]]; then
  export SOURCE_PATH="$(${READLINK_BIN} -f $(dirname ${0})/..)"
else
  export SOURCE_PATH="$(${READLINK_BIN} -f "${SOURCE_PATH}")"
fi

if [[ -z "${BINARY_PATH}" ]]; then
  export BINARY_PATH="${SOURCE_PATH}/bin"
else
  export BINARY_PATH="$(${READLINK_BIN} -f "${BINARY_PATH}")/bin"
fi

# The `go <cmd>` commands requires to see the target repository to be part of a
# Go workspace. Thus, if we are not yet in a Go workspace, let's create one
# temporarily by using symbolic links.
if [[ "${SOURCE_PATH}" != *"src/github.com/23technologies/gardener-extension-provider-hcloud" ]]; then
  SOURCE_SYMLINK_PATH="${SOURCE_PATH}/tmp/src/github.com/23technologies/gardener-extension-provider-hcloud"

  if [[ -d "${SOURCE_PATH}/tmp" && $TEST_CLEANUP == true ]]; then
    rm -rf "${SOURCE_PATH}/tmp"
  fi

  if [[ ! -d "${SOURCE_PATH}/tmp" ]]; then
    mkdir -p "${SOURCE_PATH}/tmp/src/github.com/23technologies"
    ln -s "${SOURCE_PATH}" "${SOURCE_SYMLINK_PATH}"
  fi

  cd "${SOURCE_SYMLINK_PATH}"

  export GOPATH="${SOURCE_PATH}/tmp"
  export GOBIN="${SOURCE_PATH}/tmp/bin"
  export PATH="${GOBIN}:${PATH}"
fi

LD_FLAGS="${LD_FLAGS:-$(${GOPATH}/vendor/github.com/gardener/gardener/hack/get-build-ld-flags.sh)}"

# If no LOCAL_BUILD environment variable is set, we configure the `go build` command
# to build for linux OS, amd64 architectures and without CGO enablement.
if [[ -z "$LOCAL_BUILD" ]]; then
  CGO_ENABLED=0 GOOS=$(go env GOOS) GOARCH=$(go env GOARCH) go build \
    -a \
    -v \
    -ldflags "$LD_FLAGS" \
    -o ${BINARY_PATH}/rel/gardener-extension-provider-hcloud \
    cmd/gardener-extension-provider-hcloud/main.go

  #CGO_ENABLED=0 GOOS=$(go env GOOS) GOARCH=$(go env GOARCH) go build \
  #  -a \
  #  -v \
  #  -ldflags "$LD_FLAGS" \
  #  -o ${BINARY_PATH}/rel/gardener-extension-validator-hcloud \
  #  cmd/gardener-extension-validator-hcloud/main.go
# If the LOCAL_BUILD environment variable is set, we simply run `go build`.
else
  GOOS=$(go env GOOS) GOARCH=$(go env GOARCH) go build \
    -v \
    -ldflags "$LD_FLAGS" \
    -o ${BINARY_PATH}/gardener-extension-provider-hcloud \
    cmd/gardener-extension-provider-hcloud/main.go

  #GOOS=$(go env GOOS) GOARCH=$(go env GOARCH) go build \
  #  -v \
  #  -ldflags "$LD_FLAGS" \
  #  -o ${BINARY_PATH}/gardener-extension-validator-hcloud \
  #  cmd/gardener-extension-validator-hcloud/main.go
fi

echo "Build script finished"
