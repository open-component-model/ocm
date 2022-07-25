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
ifeq ($(OS),Windows_NT)
	detected_OS := Windows
else
	detected_OS := $(shell sh -c 'uname 2>/dev/null || echo Unknown')
endif

deps := 
# TODO exact versions to compare
gSED := $(shell (sed --version 2>/dev/null || echo 0.0) | head -n 1 | sed 's/.*(GNU sed) \([0-9\.]*\).*/\1/')
ifeq ("v$(gSED)","v0.0")
	deps += $(detected_OS)_sed
endif
gTAR := $(shell (tar --version 2>/dev/null || echo 0.0) | head -n 1 | sed 's/.*(GNU tar) \([0-9\.]*\).*/\1/')
ifeq ("v$(gTAR)","v0.0")
	deps += $(detected_OS)_tar
endif
gCOREUTILS := $(shell (basename --version 2>/dev/null || echo 0.0) | head -n 1 | sed 's/.*(GNU coreutils) \([0-9\.]*\).*/\1/')
ifeq ("v$(gCOREUTILS)","v0.0")
	deps += $(detected_OS)_coreutils
endif
gGREP := $(shell (grep --version 2>/dev/null || echo 0.0) | head -n 1 | sed 's/.*(GNU grep) \([0-9\.]*\).*/\1/')
ifeq ("v$(gGREP)","v0.0")
	deps += $(detected_OS)_grep
endif
JQ := $(shell (jq --version 2>/dev/null || echo 0.0) | sed 's/.*-\([0-9\.]*\).*/\1/')
ifeq ("v$(JQ)","v0.0")
	deps += $(detected_OS)_jq
endif

GOLANGCILINT_VERSION := "v1.47.0"
GOLANGCILINTV := $(shell (golangci-lint --version 2>/dev/null || echo 0.0.0) | sed 's/.*v\([0-9\.]*\) .*/\1/')
ifneq ("v$(GOLANGCILINTV)",$(GOLANGCILINT_VERSION))
  deps += golangci-lint
endif
GO_BINDATA_VERSION := "v3.1.3"
GO_BINDATA := $(shell (go-bindata -version 2>/dev/null || echo 0.0.0) | head -n 1 | sed 's/.*go-bindata \([0-9\.]*\).*/\1/')
ifneq ("v$(GO_BINDATA)",$(GO_BINDATA_VERSION))
	deps += go-bindata
endif


build: ${SOURCES}
	go build -ldflags "-s -w \
		-X github.com/open-component-model/ocm/pkg/version.gitVersion=$(EFFECTIVE_VERSION) \
		-X github.com/open-component-model/ocm/pkg/version.gitTreeState=$(shell [ -z git status --porcelain 2>/dev/null ] && echo clean || echo dirty) \
		-X github.com/open-component-model/ocm/pkg/version.gitCommit=$(shell git rev-parse --verify HEAD) \
		-X github.com/open-component-model/ocm/pkg/version.buildDate=$(shell date --rfc-3339=seconds | sed 's/ /T/')" \
		./cmds/ocm

.PHONY: install-requirements
install-requirements: $(deps) $(GOPATH)/bin/goimports
#	@$(REPO_ROOT)/hack/install-requirements.sh

.PHONY: golangci-lint
golangci-lint:
	go install -v github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCILINT_VERSION)

.PHONY: go-bindata
go-bindata:
	go install -v  github.com/go-bindata/go-bindata/v3/...@$(GO_BINDATA_VERSION)

$(GOPATH)/bin/goimports:
	go install -v golang.org/x/tools/cmd/goimports@latest

Darwin_sed: Darwin
	$(info -> GNU sed is missing)
	$(info -  brew install gnu-sed)
	$(info -  export PATH=/usr/local/opt/gnu-sed/libexec/gnubin:$$PATH)

Darwin_tar: Darwin
	$(info -> GNU tar is missing)
	$(info -  brew install gnu-tar)
	$(info -  export PATH=/usr/local/opt/gnu-tar/libexec/gnubin:$$PATH)

Darwin_grep: Darwin
	$(info -> GNU grep is missing)
	$(info -  brew install grep)
	$(info -  export PATH=/usr/local/opt/grep/libexec/gnubin:$$PATH)

Darwin_coreutils: Darwin
	$(info -> GNU Core Utils are missing)
	$(info -  brew install coreutils)
	$(info -  export PATH=/usr/local/opt/coreutils/libexec/gnubin:$$PATH)

Darwin_jq: Darwin
	$(info -> jq is missing)
	$(info -  brew install jq)

.PHONY: Darwin
Darwin:
	$(info You are running in a MAC OS environment!)
	$(info Please make sure you have installed the following tools.)
	$(info Please allow all GNU tools to be used without their "g" prefix.)

.PHONY: prepare
prepare: generate format test

.PHONY: format
format:
	@$(REPO_ROOT)/hack/format.sh $(REPO_ROOT)/pkg $(REPO_ROOT)/cmds/ocm

.PHONY: check
check:
	@$(REPO_ROOT)/hack/check.sh --golangci-lint-config=./.golangci.yaml $(REPO_ROOT)/cmd/... $(REPO_ROOT)/pkg/... $(REPO_ROOT)/ociclient/...

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

