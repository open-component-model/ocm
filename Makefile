# SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Gardener contributors.
#
# SPDX-License-Identifier: Apache-2.0

REPO_ROOT                                      := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
VERSION                                        := $(shell cat $(REPO_ROOT)/VERSION)
EFFECTIVE_VERSION                              := $(VERSION)+$(shell git rev-parse HEAD)

REGISTRY                                       := ghcr.io/mandelsoft/ocm
COMPONENT_CLI_IMAGE_REPOSITORY                 := $(REGISTRY)/cli

SOURCES := $(shell go list -f '{{$$I:=.Dir}}{{range .GoFiles }}{{$$I}}/{{.}} {{end}}' ./... )
GOPATH                                         := $(shell go env GOPATH)

build: ${SOURCES}
	go build -ldflags "-s -w \
		-X github.com/open-component-model/ocm/pkg/version.gitVersion=$(EFFECTIVE_VERSION) \
		-X github.com/open-component-model/ocm/pkg/version.gitTreeState=$(shell [ -z git status --porcelain 2>/dev/null ] && echo clean || echo dirty) \
		-X github.com/open-component-model/ocm/pkg/version.gitCommit=$(shell git rev-parse --verify HEAD) \
		-X github.com/open-component-model/ocm/pkg/version.buildDate=$(shell date --rfc-3339=seconds | sed 's/ /T/')" \
		./cmds/ocm

.PHONY: install-requirements
install-requirements:
	@make -C hack $@

.PHONY: prepare
prepare: generate format test check

.PHONY: format
format:
	@$(REPO_ROOT)/hack/format.sh $(REPO_ROOT)/pkg $(REPO_ROOT)/cmds/ocm

.PHONY: check
check:
	@$(REPO_ROOT)/hack/check.sh --golangci-lint-config=./.golangci.yaml $(REPO_ROOT)/cmds/ocm/... $(REPO_ROOT)/pkg/...

.PHONY: test
test:
	@go test $(REPO_ROOT)/cmds/ocm/... $(REPO_ROOT)/pkg/...

.PHONY: generate
generate:
	@$(REPO_ROOT)/hack/generate.sh $(REPO_ROOT)/pkg... $(REPO_ROOT)/cmds/ocm/...

.PHONY: verify
verify: check

.PHONY: all
all: generate format test verify build


.PHONY: install
install:
	@EFFECTIVE_VERSION=$(EFFECTIVE_VERSION) ./hack/install.sh

.PHONY: cross-build
cross-build:
	@EFFECTIVE_VERSION=$(EFFECTIVE_VERSION) ./hack/cross-build.sh

