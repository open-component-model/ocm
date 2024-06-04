NAME                                           := ocm
REPO_ROOT                                      := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
GITHUBORG                                      ?= open-component-model
OCMREPO                                        ?= ghcr.io/$(GITHUBORG)/ocm
VERSION                                        := $(shell go run pkg/version/generate/release_generate.go print-rc-version $(CANDIDATE))
EFFECTIVE_VERSION                              := $(VERSION)+$(shell git rev-parse HEAD)
GIT_TREE_STATE                                 := $(shell [ -z "$$(git status --porcelain 2>/dev/null)" ] && echo clean || echo dirty)
COMMIT                                         := $(shell git rev-parse --verify HEAD)

CONTROLLER_TOOLS_VERSION ?= v0.14.0
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen

CREDS    ?=
OCM      := go run $(REPO_ROOT)/cmds/ocm $(CREDS)
CTF_TYPE ?= tgz

GEN := $(REPO_ROOT)/gen

SOURCES := $(shell go list -f '{{$$I:=.Dir}}{{range .GoFiles }}{{$$I}}/{{.}} {{end}}' ./... )
GOPATH                                         := $(shell go env GOPATH)

NOW         := $(shell date -u +%FT%T%z)
BUILD_FLAGS := "-s -w \
 -X github.com/open-component-model/ocm/pkg/version.gitVersion=$(EFFECTIVE_VERSION) \
 -X github.com/open-component-model/ocm/pkg/version.gitTreeState=$(GIT_TREE_STATE) \
 -X github.com/open-component-model/ocm/pkg/version.gitCommit=$(COMMIT) \
 -X github.com/open-component-model/ocm/pkg/version.buildDate=$(NOW)"

COMPONENTS ?= ocmcli helminstaller demoplugin ecrplugin helmdemo subchartsdemo

.PHONY: build
build: ${SOURCES}
	mkdir -p bin
	go build ./pkg/...
	go build ./examples/...
	CGO_ENABLED=0 go build -ldflags $(BUILD_FLAGS) -o bin/ocm ./cmds/ocm
	CGO_ENABLED=0 go build -ldflags $(BUILD_FLAGS) -o bin/helminstaller ./cmds/helminstaller
	CGO_ENABLED=0 go build -ldflags $(BUILD_FLAGS) -o bin/demo ./cmds/demoplugin
	CGO_ENABLED=0 go build -ldflags $(BUILD_FLAGS) -o bin/ecrplugin ./cmds/ecrplugin


.PHONY: install-requirements
install-requirements:
	@make -C hack $@

.PHONY: prepare
prepare: generate format generate-deepcopy build test check

.PHONY: format
format:
	@$(REPO_ROOT)/hack/format.sh $(REPO_ROOT)/pkg $(REPO_ROOT)/cmds/ocm $(REPO_ROOT)/cmds/helminstaller $(REPO_ROOT)/cmds/ecrplugin $(REPO_ROOT)/cmds/demoplugin

.PHONY: check
check:
	@$(REPO_ROOT)/hack/check.sh --golangci-lint-config=./.golangci.yaml $(REPO_ROOT)/cmds/ocm $(REPO_ROOT)/cmds/helminstaller/... $(REPO_ROOT)/cmds/ecrplugin/... $(REPO_ROOT)/cmds/demoplugin/... $(REPO_ROOT)/pkg/...

.PHONY: force-test
force-test:
	@go test -parallel=1 --count=1 $(REPO_ROOT)/cmds/ocm $(REPO_ROOT)/cmds/helminstaller $(REPO_ROOT)/cmds/ocm/... $(REPO_ROOT)/cmds/ecrplugin/... $(REPO_ROOT)/cmds/demoplugin/... $(REPO_ROOT)/pkg/...

.PHONY: test
test:
	@echo "> Run Unit Tests"
	@go test  ./examples/lib/... $(REPO_ROOT)/cmds/ocm/... $(REPO_ROOT)/cmds/demoplugin/... $(REPO_ROOT)/pkg/...

.PHONY: test-all
test-all: install-requirements
	@echo "> Run All Tests"
	@go test --tags=integration ./examples/lib/... $(REPO_ROOT)/cmds/ocm/... $(REPO_ROOT)/cmds/demoplugin/... $(REPO_ROOT)/pkg/...

.PHONY: generate
generate:
	@$(REPO_ROOT)/hack/generate.sh $(REPO_ROOT)/pkg/... $(REPO_ROOT)/cmds/ocm/... $(REPO_ROOT)/cmds/helminstaller/... $(REPO_ROOT)/examples/...

.PHONY: generate-deepcopy
generate-deepcopy: controller-gen
	$(CONTROLLER_GEN) object paths=./pkg/contexts/ocm/compdesc/versions/... paths=./pkg/contexts/ocm/compdesc/meta/...

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary.
$(CONTROLLER_GEN): $(LOCALBIN)
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)

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

.PHONY: info
info:
	@if [ -n "$(CANDIDATE)" ]; then echo "CANDIDATE     = $(CANDIDATE)"; fi
	@echo "VERSION       = $(VERSION)"
	@echo "EFFECTIVE     = $(EFFECTIVE_VERSION)"
	@echo "COMMIT        = $(COMMIT)"
	@echo "GIT_TREE_STATE= $(GIT_TREE_STATE)"
	@echo "COMPONENTS    = $(COMPONENTS)"

$(GEN)/.exists:
	@mkdir -p $(GEN)
	@touch $@

.PHONY: components
components: $(GEN)/.comps

$(GEN)/.comps: $(GEN)/.exists
	@rm -rf "$(GEN)"/ctf
	@for i in $(COMPONENTS); do \
       echo "building component $$i..."; \
       (cd components/$$i; make ctf;); \
    done
	@touch $@

.PHONY: ctf
ctf: $(GEN)/ctf

$(GEN)/ctf: $(GEN)/.exists $(GEN)/.comps
	@rm -rf "$(GEN)"/ctf
	@for i in $(COMPONENTS); do \
      echo "transfering component $$i..."; \
	  echo $(OCM) transfer cv  --type $(CTF_TYPE) -V $(GEN)/$$i/ctf $(GEN)/ctf; \
	  $(OCM) transfer cv  --type $(CTF_TYPE) -V $(GEN)/$$i/ctf $(GEN)/ctf; \
	done
	@touch $@

.PHONY: push
push: $(GEN)/ctf $(GEN)/.push.$(NAME)

$(GEN)/.push.$(NAME): $(GEN)/ctf
	$(OCM) transfer ctf -f $(GEN)/ctf $(OCMREPO)
	@touch $@

.PHONY: plain-push
plain-push: $(GEN)
	$(OCM) transfer ctf -f $(GEN)/ctf $(OCMREPO)

.PHONY: plain-ctf
plain-ctf: $(GEN)
	@rm -rf "$(GEN)"/ctf
	@for i in $(COMPONENTS); do \
       echo "transfering component $$i..."; \
       echo $(OCM) transfer cv  --type $(CTF_TYPE) -V $(GEN)/$$i/ctf $(GEN)/ctf; \
       $(OCM) transfer cv  --type $(CTF_TYPE) -V $(GEN)/$$i/ctf $(GEN)/ctf; \
     done

.PHONY: clean
clean:
	rm -rf "$(GEN)"
