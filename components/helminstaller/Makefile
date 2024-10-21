NAME      := helminstaller
PROVIDER  ?= ocm.software
GITHUBORG ?= open-component-model
IMAGE     := $(NAME)
COMPONENT := $(PROVIDER)/toi/installers/$(NAME)
OCMREPO   ?= ghcr.io/$(GITHUBORG)/ocm
MULTI     ?= true
PLATFORMS ?= linux/amd64 linux/arm64

REPO_ROOT                                      := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))/../..
VERSION                                        := $(shell go run ../../api/version/generate/release_generate.go print-rc-version $(CANDIDATE))
COMMIT                                         := $(shell git rev-parse --verify HEAD)
EFFECTIVE_VERSION                              := $(VERSION)-$(COMMIT)
GIT_TREE_STATE                                 := $(shell [ -z "$$(git status --porcelain 2>/dev/null)" ] && echo clean || echo dirty)
PLATFORM                                       := $(shell go env GOOS)/$(shell go env GOARCH)
CACHE_DIR									  := $(shell go env GOCACHE)
MOD_CACHE_DIR								  := $(shell go env GOMODCACHE)

CREDS ?=
# Define the path to the binary
OCM_BIN = $(REPO_ROOT)/bin/ocm

# Rule to build the binary if it doesn't exist or if the source code has changed
$(OCM_BIN): $(REPO_ROOT)/cmds/ocm/main.go
	mkdir -p $(REPO_ROOT)/bin
	go build -ldflags $(BUILD_FLAGS) -o $(OCM_BIN) $(REPO_ROOT)/cmds/ocm

# Use the binary for the OCM command
OCM = $(OCM_BIN) $(CREDS)

GEN = $(REPO_ROOT)/gen/$(NAME)

NOW         := $(shell date -u +%FT%T%z)
BUILD_FLAGS := "-s -w \
 -X ocm.software/ocm/api/version.gitVersion=$(EFFECTIVE_VERSION) \
 -X ocm.software/ocm/api/version.gitTreeState=$(GIT_TREE_STATE) \
 -X ocm.software/ocm/api/version.gitCommit=$(COMMIT) \
 -X ocm.software/ocm/api/version.buildDate=$(NOW)"

CMDSRCS=$(shell find $(REPO_ROOT)/cmds/$(NAME) -type f)
OCMSRCS=$(shell find $(REPO_ROOT)/pkg -type f) $(REPO_ROOT)/go.*

ATTRIBUTES = VERSION="$(VERSION)" COMMIT="$(COMMIT)" IMAGE="$(IMAGE):$(VERSION)" PLATFORMS="$(PLATFORMS)" MULTI=$(MULTI)

ifeq ($(MULTI),true)
FLAGSUF     = .multi
endif

.PHONY: ctf
ctf: $(GEN)/ctf

$(GEN)/ctf: $(GEN)/ca $(OCM_BIN)
	@rm -rf "$(GEN)/ctf"
	$(OCM) transfer ca $(GEN)/ca $(GEN)/ctf
	touch $(GEN)/ctf

.PHONY: plain-ctf
plain-ctf: $(OCM_BIN)
	$(OCM) transfer ca $(GEN)/ca $(GEN)/ctf
	touch $(GEN)/ctf

.PHONY: version
version:
	@echo $(VERSION)

.PHONY: ca
ca: $(GEN)/ca

$(GEN)/ca: $(GEN)/.exists $(GEN)/image.$(NAME)$(FLAGSUF) resources.yaml executorspec.yaml $(OCM_BIN)
	$(OCM) create ca -f $(COMPONENT) "$(VERSION)" --provider $(PROVIDER) --file $(GEN)/ca
	$(OCM) add resources --templater spiff $(GEN)/ca $(ATTRIBUTES)  resources.yaml
	@touch $(GEN)/ca


.PHONY: plain-ca
plain-ca: $(GEN)/.exists $(OCM_BIN)
	$(OCM) create ca -f $(COMPONENT) "$(VERSION)" --provider  $(PROVIDER) --file $(GEN)/ca
	$(OCM) add resources --templater spiff $(GEN)/ca $(ATTRIBUTES) resources.yaml
	@touch $(GEN)/ca

.PHONY: eval-resources
eval-resources: $(OCM_BIN)
	$(OCM) add resources --dry-run --templater spiff $(ATTRIBUTES) resources.yaml

.PHONY: build
build: $(GEN)/image.$(NAME)$(FLAGSUF)

$(GEN)/image.$(NAME): $(GEN)/.exists Dockerfile $(CMDSRCS) $(OCMSRCS)
	docker buildx build -t $(IMAGE):$(VERSION) --platform $(PLATFORM) --file Dockerfile $(REPO_ROOT) \
          --build-arg COMMIT=$(COMMIT) \
          --build-arg CACHE_DIR=$(CACHE_DIR) \
          --build-arg MOD_CACHE_DIR=$(MOD_CACHE_DIR) \
          --build-arg EFFECTIVE_VERSION=$(EFFECTIVE_VERSION) \
          --build-arg GIT_TREE_STATE=$(GIT_TREE_STATE) \
          --cache-from type=local,src=$(CACHE_DIR) \
          --cache-from type=local,src=$(MOD_CACHE_DIR); \
	@touch $(GEN)/image.$(NAME)

push-image:
	docker tag $(IMAGE):$(VERSION) $(OCMREPO)/$(COMPONENT)/$(NAME):$(VERSION)
	docker push $(OCMREPO)/$(COMPONENT)/$(NAME):$(VERSION)

.PHONY: multi
multi: $(GEN)/image.$(NAME).multi

$(GEN)/image.$(NAME).multi: $(GEN)/.exists Dockerfile $(CMDSRCS) $(OCMSRCS)
	for i in $(PLATFORMS); do \
	tag=$$(echo $$i | sed -e s:/:-:g); \
        echo building platform $$i; \
	docker buildx build --load -t $(IMAGE):$(VERSION)-$$tag --platform $$i --file Dockerfile $(REPO_ROOT) \
          --build-arg COMMIT=$(COMMIT) \
          --build-arg CACHE_DIR=$(CACHE_DIR) \
          --build-arg MOD_CACHE_DIR=$(MOD_CACHE_DIR) \
          --build-arg EFFECTIVE_VERSION=$(EFFECTIVE_VERSION) \
          --build-arg GIT_TREE_STATE=$(GIT_TREE_STATE) \
          --cache-from type=local,src=$(CACHE_DIR) \
          --cache-from type=local,src=$(MOD_CACHE_DIR); \
	done
	@touch $(GEN)/image.$(NAME).multi

.PHONY: push
push: $(GEN)/ctf $(GEN)/push.$(NAME)

$(GEN)/push.$(NAME): $(GEN)/ctf $(OCM_BIN)
	$(OCM) transfer ctf -f $(GEN)/ctf $(OCMREPO)
	@touch $(GEN)/push.$(NAME)

.PHONY: plain-push
plain-push: $(GEN)/.exists $(OCM_BIN)
	$(OCM) transfer ctf -f $(GEN)/ctf $(OCMREPO)
	@touch $(GEN)/push.$(NAME)

.PHONY: transport
transport: $(OCM_BIN)
ifneq ($(TARGETREPO),)
	$(OCM) transfer component -Vr  $(OCMREPO)//$(COMPONENT):$(VERSION) $(TARGETREPO)
endif

$(GEN)/.exists:
	@mkdir -p $(GEN)
	@touch $@

.PHONY: info
info:
	@echo "ROOT:        $(REPO_ROOT)"
	@echo "VERSION:     $(VERSION)"
	@echo "COMMIT:      $(COMMIT)"
	@echo "GITHUBORG:   $(GITHUBORG)"
	@echo "PATFORM:     $(PLATFORM)"

.PHONY: describe
describe: $(GEN)/ctf
	ocm get resources --lookup $(OCMREPO) -r -o treewide $(GEN)/ctf

.PHONY: descriptor
descriptor: $(GEN)/ctf
	ocm get component -S v3alpha1 -o yaml $(GEN)/ctf

.PHONY: setup
setup:
	docker run --userns=host --privileged --rm docker/binfmt:a7996909642ee92942dcd6cff44b9b95f08dad64

.PHONY: clean
clean:
	rm -rf $(GEN)
