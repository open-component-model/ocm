# SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
#
# SPDX-License-Identifier: Apache-2.0

NAME                                           := ocm
REPO_ROOT                                      := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
VERSION                                        := $(shell go run pkg/version/generate/release_generate.go print-version)
GITHUBORG                                      ?= open-component-model
OCMREPO                                        ?= ghcr.io/$(GITHUBORG)/ocm
EFFECTIVE_VERSION                              := $(VERSION)+$(shell git rev-parse HEAD)
GIT_TREE_STATE                                 := $(shell [ -z "$$(git status --porcelain 2>/dev/null)" ] && echo clean || echo dirty)
COMMIT                                         := $(shell git rev-parse --verify HEAD)

CREDS := $(shell $(REPO_ROOT)/hack/githubcreds.sh)
OCM := go run $(REPO_ROOT)/cmds/ocm $(CREDS)

GEN := $(REPO_ROOT)/gen

SOURCES := $(shell go list -f '{{$$I:=.Dir}}{{range .GoFiles }}{{$$I}}/{{.}} {{end}}' ./... )
GOPATH                                         := $(shell go env GOPATH)

NOW         := $(shell date -u +%FT%T%z)
BUILD_FLAGS := "-s -w \
 -X github.com/open-component-model/ocm/pkg/version.gitVersion=$(EFFECTIVE_VERSION) \
 -X github.com/open-component-model/ocm/pkg/version.gitTreeState=$(GIT_TREE_STATE) \
 -X github.com/open-component-model/ocm/pkg/version.gitCommit=$(COMMIT) \
 -X github.com/open-component-model/ocm/pkg/version.buildDate=$(NOW)"

build: ${SOURCES}
	mkdir -p bin
	go build -ldflags $(BUILD_FLAGS) -o bin/ocm ./cmds/ocm
	go build -ldflags $(BUILD_FLAGS) -o bin/helminstaller ./cmds/helminstaller
	go build -ldflags $(BUILD_FLAGS) -o bin/demo ./cmds/demoplugin
	go build -ldflags $(BUILD_FLAGS) -o bin/ecrplugin ./cmds/ecrplugin


.PHONY: install-requirements
install-requirements:
	@make -C hack $@

.PHONY: prepare
prepare: generate format build test check

.PHONY: format
format:
	@$(REPO_ROOT)/hack/format.sh $(REPO_ROOT)/pkg $(REPO_ROOT)/cmds/ocm $(REPO_ROOT)/cmds/helminstaller $(REPO_ROOT)/cmds/ecrplugin $(REPO_ROOT)/cmds/demoplugin

.PHONY: check
check:
	@$(REPO_ROOT)/hack/check.sh --golangci-lint-config=./.golangci.yaml $(REPO_ROOT)/cmds/ocm $(REPO_ROOT)/cmds/helminstaller/... $(REPO_ROOT)/cmds/ecrplugin/... $(REPO_ROOT)/cmds/demoplugin/... $(REPO_ROOT)/pkg/...

.PHONY: force-test
force-test:
	@go test --count=1 $(REPO_ROOT)/cmds/ocm $(REPO_ROOT)/cmds/helminstaller $(REPO_ROOT)/cmds/ocm/... $(REPO_ROOT)/cmds/ecrplugin/... $(REPO_ROOT)/cmds/demoplugin/... $(REPO_ROOT)/pkg/...

.PHONY: test
test:
	@echo "> Test"
	@go test $(REPO_ROOT)/cmds/ocm/... $(REPO_ROOT)/pkg/...

.PHONY: generate
generate:
	@$(REPO_ROOT)/hack/generate.sh $(REPO_ROOT)/pkg... $(REPO_ROOT)/cmds/ocm/... $(REPO_ROOT)/cmds/helminst/...

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

##@ Build Dependencies

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

.PHONY: generate-license
generate-license:
	for f in $(shell find . -name "*.go" -o -name "*.sh"); do \
		reuse addheader -r --copyright="SAP SE or an SAP affiliate company and Open Component Model contributors." --license="Apache-2.0" $$f --skip-unrecognised; \
	done


$(GEN)/.exists:
	@mkdir -p $(GEN)
	@touch $@

.PHONY: components
components: $(GEN)/.comps

$(GEN)/.comps:
	@echo Helminstaller; cd components/helminstaller; make ctf
	@echo HelmDemo; cd components/helmdemo; make ctf
	@echo OCMCLI; cd components/ocmcli; make ctf
	touch $@

.PHONY: ctf
ctf: $(GEN)/ctf

$(GEN)/ctf: $(GEN)/.exists components
	@rm -rf "$(GEN)"/ctf
	$(OCM) transfer cv -V $(GEN)/helminstaller/ctf $(GEN)/ctf
	$(OCM) transfer cv -V $(GEN)/helmdemo/ctf $(GEN)/ctf
	$(OCM) transfer cv -V $(GEN)/ocmcli/ctf $(GEN)/ctf
	touch $@

.PHONY: push
push: $(GEN)/ctf $(GEN)/.push.$(NAME)

$(GEN)/.push.$(NAME): $(GEN)/ctf
	$(OCM) transfer ctf -f $(GEN)/ctf $(OCMREPO)
	@touch $@

.PHONY: plain-ctf
plain-ctf: $(GEN)
	@rm -rf "$(GEN)"/ctf
	$(OCM) transfer cv -V $(GEN)/helminstaller/ctf $(GEN)/ctf
	$(OCM) transfer cv -V $(GEN)/helmdemo/ctf $(GEN)/ctf
	$(OCM) transfer cv -V $(GEN)/ocmcli/ctf $(GEN)/ctf

