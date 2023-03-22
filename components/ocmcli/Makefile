NAME      = ocmcli
CMD       = ocm
PROVIDER  ?= ocm.software
GITHUBORG  ?= open-component-model
COMPONENT = $(PROVIDER)/$(NAME)
OCMREPO   ?= ghcr.io/$(GITHUBORG)/ocm
PLATFORMS = linux/amd64 linux/arm64 darwin/arm64 darwin/amd64 windows/amd64

REPO_ROOT                                      := $(dir $(realpath $(lastword $(MAKEFILE_LIST))))../..
GIT_TREE_STATE                                 = $(shell [ -z "$$(git status --porcelain 2>/dev/null)" ] && echo clean || echo dirty)
VERSION                                        = $(shell go run ../../pkg/version/generate/release_generate.go print-rc-version $(CANDIDATE))
COMMIT                                         = $(shell git rev-parse HEAD)
EFFECTIVE_VERSION                              = $(VERSION)+$(COMMIT)

CMDSRCS=$(shell find $(REPO_ROOT)/cmds/$(CMD) -type f) Makefile
OCMSRCS=$(shell find $(REPO_ROOT)/pkg -type f) $(REPO_ROOT)/go.*

CREDS := $(shell $(REPO_ROOT)/hack/githubcreds.sh)
OCM = go run $(REPO_ROOT)/cmds/ocm $(CREDS)

GEN = $(REPO_ROOT)/gen/$(shell basename $(realpath .))

NOW         := $(shell date -u +%FT%T%z)
BUILD_FLAGS := "-s -w \
 -X github.com/open-component-model/ocm/pkg/version.gitVersion=$(EFFECTIVE_VERSION) \
 -X github.com/open-component-model/ocm/pkg/version.gitTreeState=$(GIT_TREE_STATE) \
 -X github.com/open-component-model/ocm/pkg/version.gitCommit=$(COMMIT) \
 -X github.com/open-component-model/ocm/pkg/version.buildDate=$(NOW)"


.PHONY: build
build: $(GEN)/build

$(GEN)/build: $(CMDSRCS) $(OCMSRCS)
	@for i in $(PLATFORMS); do \
    tag=$$(echo $$i | sed -e s:/:-:g); \
    echo GOARCH=$$(basename $$i) GOOS=$$(dirname $$i) go build -ldflags $(BUILD_FLAGS) -o $(GEN)/$(NAME).$$tag ../../cmds/$(CMD); \
    GOARCH=$$(basename $$i) GOOS=$$(dirname $$i) go build -ldflags $(BUILD_FLAGS) -o $(GEN)/$(NAME).$$tag ../../cmds/$(CMD); \
    done
	@touch $(GEN)/build


.PHONY: ctf
ctf: $(GEN)/ctf

$(GEN)/ctf: $(GEN)/ca.done
	@rm -rf "$(GEN)/ctf"
	$(OCM) transfer ca $(GEN)/ca $(GEN)/ctf
	touch $(GEN)/ctf

.PHONY: version
version:
	@echo $(VERSION)

.PHONY: ca
ca: $(GEN)/ca.done

$(GEN)/ca.done: $(GEN)/.exists $(GEN)/build resources.yaml $(CHARTSRCS) Makefile
	$(OCM) create ca -f $(COMPONENT) "$(VERSION)" --provider $(PROVIDER) --file $(GEN)/ca
	$(OCM) add resources --templater=spiff --file $(GEN)/ca VERSION="$(VERSION)" NAME="$(NAME)" COMMIT="$(COMMIT)" GEN="$(GEN)" PLATFORMS="$(PLATFORMS)" resources.yaml
	@touch $(GEN)/ca.done

.PHONY: push
push: $(GEN)/ctf $(GEN)/push.$(NAME)

$(GEN)/push.$(NAME): $(GEN)/ctf
	$(OCM) transfer ctf -f $(GEN)/ctf $(OCMREPO)
	@touch $(GEN)/push.$(NAME)

.PHONY: plain-push
plain-push: $(GEN)
	$(OCM) transfer ctf -f $(GEN)/ctf $(OCMREPO)
	@touch $(GEN)/push.$(NAME)

.PHONY: transport
transport:
ifneq ($(TARGETREPO),)
	$(OCM) transfer component -Vc  $(OCMREPO)//$(COMPONENT):$(VERSION) $(TARGETREPO)
endif

$(GEN)/.exists:
	@mkdir -p $(GEN)
	@touch $@

.PHONY: info
info:
	@echo "ROOT:     $(REPO_ROOT)"
	@echo "VERSION:  $(VERSION)"
	@echo "COMMIT:   $(COMMIT)"
	@echo "GIT_TREE: $(GIT_TREE_STATE)"

.PHONY: describe
describe: $(GEN)/ctf
	ocm get resources --lookup $(OCMREPO) -c -o treewide $(GEN)/ctf

.PHONY: descriptor
descriptor: $(GEN)/ctf
	ocm get component -S v3alpha1 -o yaml $(GEN)/ctf

.PHONY: clean
clean:
	rm -rf $(GEN)
