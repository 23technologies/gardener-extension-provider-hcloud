#!/usr/bin/env bash

# Copyright (c) 2019 SAP SE or an SAP affiliate company. All rights reserved.
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

# Argument options
while test $# -gt 0
do
    case "$1" in
        --clean) TEST_CLEANUP=true
            ;;
        --coverage) TEST_COVERAGE=true
            ;;
    esac
    shift
done

# For the test step concourse will set the following environment variables:
# SOURCE_PATH - path to component repository root directory.

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

if [[ ! -d "${GOPATH}/src/github.com/onsi/ginkgo/ginkgo" ]]; then
  # Install Ginkgo (test framework) to be able to execute the tests.
  echo "Fetching Ginkgo frawework"
  GO111MODULE=off go get -u github.com/onsi/ginkgo/ginkgo
  echo "Successfully fetched Ginkgo frawework"
fi

##############################################################################

function test_with_coverage() {
  local output_dir=test/output
  local coverprofile_file=coverprofile.out
  mkdir -p test/output
  ginkgo $GINKGO_COMMON_FLAGS --coverprofile ${coverprofile_file} -covermode=set -outputdir ${output_dir} ${TEST_PACKAGES}

  sed -i -e '/mode: set/d' ${output_dir}/${coverprofile_file}
  {( echo "mode: set"; cat ${output_dir}/${coverprofile_file} )} > ${output_dir}/${coverprofile_file}.temp
  mv ${output_dir}/${coverprofile_file}.temp ${output_dir}/${coverprofile_file}
  go tool cover -func ${output_dir}/${coverprofile_file}
}

###############################################################################

if [[ "${SKIP_UNIT_TESTS}" != "" ]]; then
  echo ">>>>>Skipping unit tests"
else
  echo ">>>>> Invoking unit tests"
  TEST_PACKAGES="pkg"
  GINKGO_COMMON_FLAGS="-r -timeout=1h0m0s --randomizeAllSpecs --randomizeSuites --failOnPending  --progress"

  if [[ $TEST_COVERAGE == true ]]; then
    test_with_coverage
  else
    ginkgo $GINKGO_COMMON_FLAGS ${TEST_PACKAGES}
  fi

  echo ">>>>> Finished executing unit tests"
fi

echo "CI tests have passed successfully"
