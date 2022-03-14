# SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Gardener contributors.
#
# SPDX-License-Identifier: Apache-2.0

REPO_ROOT                                      := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
VERSION                                        := $(shell cat $(REPO_ROOT)/VERSION)
EFFECTIVE_VERSION                              := $(VERSION)-$(shell git rev-parse HEAD)

REGISTRY                                       := ghcr.io/mandelsoft/ocm
COMPONENT_CLI_IMAGE_REPOSITORY                 := $(REGISTRY)/cli

.PHONY: install-requirements
build:
	go build -ldflags "-s -w \
		-X github.com/gardener/ocm/pkg/version.gitVersion=$(EFFECTIVE_VERSION) \
		-X github.com/gardener/ocm/pkg/version.gitTreeState=$(shell [ -z git status --porcelain 2>/dev/null ] && echo clean || echo dirty) \
		-X github.com/gardener/ocm/pkg/version.gitCommit=$(shell git rev-parse --verify HEAD) \
		-X github.com/gardener/ocm/pkg/version.buildDate=$(shell date --rfc-3339=seconds | sed 's/ /T/')" \
		./cmds/ocm

.PHONY: install-requirements
install-requirements:
	@go install -mod=vendor $(REPO_ROOT)/vendor/github.com/golang/mock/mockgen
	@$(REPO_ROOT)/hack/install-requirements.sh

.PHONY: revendor
revendor:
	@$(REPO_ROOT)/hack/revendor.sh

.PHONY: format
format:
	@$(REPO_ROOT)/hack/format.sh $(REPO_ROOT)/pkg $(REPO_ROOT)/cmd $(REPO_ROOT)/ociclient

.PHONY: check
check:
	@$(REPO_ROOT)/hack/check.sh --golangci-lint-config=./.golangci.yaml $(REPO_ROOT)/cmd/... $(REPO_ROOT)/pkg/... $(REPO_ROOT)/ociclient/...

.PHONY: test
test:
	@go test -mod=vendor $(REPO_ROOT)/cmds/ocm/... $(REPO_ROOT)/pkg/...

.PHONY: generate
generate:
	@$(REPO_ROOT)/hack/generate.sh $(REPO_ROOT)/pkg... $(REPO_ROOT)/cmds/ocm/...

.PHONY: verify
verify: check

.PHONY: all
all: generate format test verify build

#################################################################
# Rules related to binary build, docker image build and release #
#################################################################

.PHONY: install
install:
	@EFFECTIVE_VERSION=$(EFFECTIVE_VERSION) ./hack/install.sh

.PHONY: cross-build
cross-build:
	@EFFECTIVE_VERSION=$(EFFECTIVE_VERSION) ./hack/cross-build.sh

.PHONY: docker-images
docker-images:
	@echo "Building docker images for version $(EFFECTIVE_VERSION)"
	@docker build -t $(COMPONENT_CLI_IMAGE_REPOSITORY):$(EFFECTIVE_VERSION) -f Dockerfile --target cli .

.PHONY: docker-images
docker-push:
	@echo "Pushing docker images for version $(EFFECTIVE_VERSION) to registry $(REGISTRY)"
	@if ! docker images $(COMPONENT_CLI_IMAGE_REPOSITORY) | awk '{ print $$2 }' | grep -q -F $(EFFECTIVE_VERSION); then echo "$(COMPONENT_CLI_IMAGE_REPOSITORY) version $(EFFECTIVE_VERSION) is not yet built. Please run 'make docker-images'"; false; fi
	@docker push $(COMPONENT_CLI_IMAGE_REPOSITORY):$(EFFECTIVE_VERSION)


