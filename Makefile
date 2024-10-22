NAME                                           := ocm
REPO_ROOT                                      := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
GITHUBORG                                      ?= open-component-model
OCMREPO                                        ?= ghcr.io/$(GITHUBORG)/ocm
VERSION                                        := $(shell go run api/version/generate/release_generate.go print-rc-version $(CANDIDATE))
EFFECTIVE_VERSION                              := $(VERSION)+$(shell git rev-parse HEAD)
GIT_TREE_STATE                                 := $(shell [ -z "$$(git status --porcelain 2>/dev/null)" ] && echo clean || echo dirty)
COMMIT                                         := $(shell git rev-parse --verify HEAD)

CONTROLLER_TOOLS_VERSION ?= v0.14.0
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen

PLATFORMS = windows/amd64 darwin/arm64 darwin/amd64 linux/amd64 linux/arm64

CREDS    ?=
OCM      := go run $(REPO_ROOT)/cmds/ocm $(CREDS)
CTF_TYPE ?= tgz

GEN := $(REPO_ROOT)/gen

SOURCES := $(shell go list -f '{{$$I:=.Dir}}{{range .GoFiles }}{{$$I}}/{{.}} {{end}}' ./... )
GOPATH                                         := $(shell go env GOPATH)

NOW         := $(shell date -u +%FT%T%z)
BUILD_FLAGS := "-s -w \
 -X ocm.software/ocm/api/version.gitVersion=$(EFFECTIVE_VERSION) \
 -X ocm.software/ocm/api/version.gitTreeState=$(GIT_TREE_STATE) \
 -X ocm.software/ocm/api/version.gitCommit=$(COMMIT) \
 -X ocm.software/ocm/api/version.buildDate=$(NOW)"

COMPONENTS ?= ocmcli helminstaller demoplugin ecrplugin helmdemo subchartsdemo

.PHONY: build
build: ${SOURCES}
	mkdir -p bin
	go build ./api/...
	go build ./examples/...
	CGO_ENABLED=0 go build -ldflags $(BUILD_FLAGS) -o bin/ocm ./cmds/ocm
	CGO_ENABLED=0 go build -ldflags $(BUILD_FLAGS) -o bin/helminstaller ./cmds/helminstaller
	CGO_ENABLED=0 go build -ldflags $(BUILD_FLAGS) -o bin/demo ./cmds/demoplugin
	CGO_ENABLED=0 go build -ldflags $(BUILD_FLAGS) -o bin/cliplugin ./cmds/cliplugin
	CGO_ENABLED=0 go build -ldflags $(BUILD_FLAGS) -o bin/ecrplugin ./cmds/ecrplugin


build-platforms: $(GEN)/.exists $(SOURCES)
	@for i in $(PLATFORMS); do \
    echo GOARCH=$$(basename $$i) GOOS=$$(dirname $$i); \
    GOARCH=$$(basename $$i) GOOS=$$(dirname $$i) CGO_ENABLED=0 go build ./cmds/ocm ./cmds/helminstaller ./cmds/ecrplugin; \
    done

.PHONY: install-requirements
install-requirements:
	@make -C hack $@

.PHONY: prepare
prepare: generate format generate-deepcopy build test check

EFFECTIVE_DIRECTORIES := $(REPO_ROOT)/cmds/ocm/... $(REPO_ROOT)/cmds/helminstaller/... $(REPO_ROOT)/cmds/ecrplugin/... $(REPO_ROOT)/cmds/demoplugin/... $(REPO_ROOT)/cmds/cliplugin/... $(REPO_ROOT)/examples/... $(REPO_ROOT)/cmds/subcmdplugin/... $(REPO_ROOT)/api/...

.PHONY: format
format:
	@$(REPO_ROOT)/hack/format.sh $(EFFECTIVE_DIRECTORIES)

.PHONY: check
check:
	@$(REPO_ROOT)/hack/check.sh --golangci-lint-config=./.golangci.yaml $(EFFECTIVE_DIRECTORIES)

.PHONY: check-and-fix
check-and-fix:
	@$(REPO_ROOT)/hack/check.sh --fix --golangci-lint-config=./.golangci.yaml $(EFFECTIVE_DIRECTORIES)

.PHONY: force-test
force-test:
	@go test --count=1 $(EFFECTIVE_DIRECTORIES)

.PHONY: test
test:
	@echo "> Run Tests"
	@go test  --tags=integration $(EFFECTIVE_DIRECTORIES)

.PHONY: unit-test
unit-test:
	@echo "> Run Unit Tests"
	@go test $(EFFECTIVE_DIRECTORIES)

.PHONY: generate
generate:
	@$(REPO_ROOT)/hack/generate.sh $(EFFECTIVE_DIRECTORIES)

.PHONY: generate-deepcopy
generate-deepcopy: controller-gen
	$(CONTROLLER_GEN) object paths=./api/ocm/compdesc/versions/... paths=./api/ocm/compdesc/meta/...

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
