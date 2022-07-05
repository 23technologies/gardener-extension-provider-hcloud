# Copyright (c) 2020 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
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

EXTENSION_PREFIX            := gardener-extension
NAME                        := provider-hcloud
ADMISSION_NAME              := admission-hcloud
REPO_ROOT                   := $(shell dirname $(realpath $(lastword ${MAKEFILE_LIST})))
HACK_DIR                    := ${REPO_ROOT}/hack
KUBECONFIG                  := dev/kubeconfig.yaml
MANAGER_CONFIG_FILE         := example/00-componentconfig.yaml
PROJECT_NAME                := 23technologies
VERSION                     := $(shell cat "${REPO_ROOT}/VERSION")
LD_FLAGS                    := "-w $(shell $(REPO_ROOT)/vendor/github.com/gardener/gardener/hack/get-build-ld-flags.sh k8s.io/component-base $(REPO_ROOT)/VERSION $(EXTENSION_PREFIX)-$(NAME))"
LEADER_ELECTION             := false
IGNORE_OPERATION_ANNOTATION := false
GARDENER_VERSION            := $(grep "gardener/gardener v" go.mod | tr "[:blank:]" "\\n" | tail -1)

WEBHOOK_CONFIG_PORT	:= 8444
WEBHOOK_CONFIG_MODE	:= url
WEBHOOK_CONFIG_URL	:= localhost:${WEBHOOK_CONFIG_PORT}
WEBHOOK_CERT_DIR    := ./example/admission-hcloud-certs
EXTENSION_NAMESPACE	:=

WEBHOOK_PARAM := --webhook-config-url=${WEBHOOK_CONFIG_URL}
ifeq (${WEBHOOK_CONFIG_MODE}, service)
  WEBHOOK_PARAM := --webhook-config-namespace=${EXTENSION_NAMESPACE}
endif

WEBHOOK_CERT_DIR=/tmp/gardener-extensions-cert

#########################################
# Rules for local development scenarios #
#########################################

.PHONY: start
start:
	@LEADER_ELECTION_NAMESPACE=garden GO111MODULE=on go run \
		-mod=vendor \
		-ldflags ${LD_FLAGS} \
		./cmd/${EXTENSION_PREFIX}-${NAME} \
		--kubeconfig=${KUBECONFIG} \
		--config-file=${MANAGER_CONFIG_FILE} \
		--ignore-operation-annotation=${IGNORE_OPERATION_ANNOTATION} \
		--leader-election=${LEADER_ELECTION} \
		--webhook-config-server-host=0.0.0.0 \
		--webhook-config-server-port=${WEBHOOK_CONFIG_PORT} \
		--webhook-config-mode=${WEBHOOK_CONFIG_MODE} \
		--gardener-version=${GARDENER_VERSION} \
		${WEBHOOK_PARAM}

.PHONY: debug
debug:
	dlv debug  ./cmd/${EXTENSION_PREFIX}-${NAME} -- \
		--kubeconfig=${KUBECONFIG} \
		--config-file=${MANAGER_CONFIG_FILE} \
		--ignore-operation-annotation=${IGNORE_OPERATION_ANNOTATION} \
		--leader-election=${LEADER_ELECTION} \
		--webhook-config-server-host=0.0.0.0 \
		--webhook-config-server-port=${WEBHOOK_CONFIG_PORT} \
		--webhook-config-mode=${WEBHOOK_CONFIG_MODE} \
		--gardener-version=${GARDENER_VERSION} \
		${WEBHOOK_PARAM}

.PHONY: start-admission
start-admission:
	@LEADER_ELECTION_NAMESPACE=garden GO111MODULE=on go run \
		-mod=vendor \
		-ldflags ${LD_FLAGS} \
		./cmd/${EXTENSION_PREFIX}-${ADMISSION_NAME} \
		--kubeconfig=dev/garden-kubeconfig.yaml \
		--leader-election=${LEADER_ELECTION} \
		--webhook-config-server-host=0.0.0.0 \
		--webhook-config-server-port=9443 \
		--health-bind-address=:8085 \
		--webhook-config-cert-dir=${WEBHOOK_CERT_DIR}

.PHONY: debug-admission
debug-admission:
	LEADER_ELECTION_NAMESPACE=garden dlv debug \
		./cmd/${EXTENSION_PREFIX}-${ADMISSION_NAME} -- \
		--leader-election=${LEADER_ELECTION} \
		--kubeconfig=dev/garden-kubeconfig.yaml \
		--webhook-config-server-host=0.0.0.0 \
		--webhook-config-server-port=9443 \
		--health-bind-address=:8085 \
		--webhook-config-cert-dir=${WEBHOOK_CERT_DIR}
#########################################
# Rules for re-vendoring
#########################################

.PHONY: revendor
revendor:
	@GO111MODULE=on go mod tidy -compat=1.17
	@GO111MODULE=on go mod vendor
	@chmod +x ${REPO_ROOT}/vendor/github.com/gardener/gardener/hack/*
	@chmod +x ${REPO_ROOT}/vendor/github.com/gardener/gardener/hack/.ci/*
	@${REPO_ROOT}/hack/update-github-templates.sh

.PHONY: update-dependencies
update-dependencies:
	@env GO111MODULE=on go get -u

#########################################
# Rules for testing
#########################################

.PHONY: test
test:
	@hack/test.sh

.PHONY: test-cov
test-cov:
	@hack/test.sh --coverage

.PHONY: test-clean
test-clean:
	@hack/test.sh --clean --coverage

#########################################
# Rules for build/release
#########################################

.PHONY: build-local
build-local:
	@env LD_FLAGS=${LD_FLAGS} LOCAL_BUILD=1 hack/build.sh

.PHONY: build
build:
	@env LD_FLAGS=${LD_FLAGS} hack/build.sh

.PHONY: clean
clean:
	@$(shell find ./example -type f -name "controller-registration.yaml" -exec rm '{}' \;)
	@${REPO_ROOT}/vendor/github.com/gardener/gardener/hack/clean.sh ./cmd/... ./pkg/... ./test/... ./tmp/

#########################################
# Rules for verification
#########################################

.PHONY: check-generate
check-generate:
	@${REPO_ROOT}/vendor/github.com/gardener/gardener/hack/check-generate.sh ${REPO_ROOT}

.PHONY: check
check:
	@${REPO_ROOT}/vendor/github.com/gardener/gardener/hack/check.sh --golangci-lint-config=./.golangci.yaml ./cmd/... ./pkg/... ./test/...
	@${REPO_ROOT}/vendor/github.com/gardener/gardener/hack/check-charts.sh ./charts

.PHONY: generate
generate:
	@${REPO_ROOT}/vendor/github.com/gardener/gardener/hack/generate.sh ./charts/... ./cmd/... ./pkg/... ./test/...

.PHONY: format
format:
	@${REPO_ROOT}/vendor/github.com/gardener/gardener/hack/format.sh ./cmd ./pkg ./test

.PHONY: verify
verify: check format test

.PHONY: install-requirements
install-requirements:
	@go install -mod=vendor ${REPO_ROOT}/vendor/github.com/ahmetb/gen-crd-api-reference-docs
	@go install -mod=vendor ${REPO_ROOT}/vendor/github.com/golang/mock/mockgen
	@go install -mod=vendor ${REPO_ROOT}/vendor/github.com/onsi/ginkgo/v2/ginkgo
	@${REPO_ROOT}/vendor/github.com/gardener/gardener/hack/install-requirements.sh

.PHONY: verify-extended
verify-extended: install-requirements check-generate check format test-clean

#########################################
# Rules for infra-cli
#########################################

.PHONY: install-infra-cli
install-infra-cli:
	@go install -mod=vendor ./cmd/infra-cli
