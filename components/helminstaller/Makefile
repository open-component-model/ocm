NAME      := helminstaller
PROVIDER  ?= ocm.software
GITHUBORG ?= open-component-model
IMAGE     := $(NAME)
COMPONENT := $(PROVIDER)/toi/installers/$(NAME)
OCMREPO   ?= ghcr.io/$(GITHUBORG)/ocm
MULTI     ?= true
IMAGE_PLATFORMS ?= linux/amd64 linux/arm64
CTF_TYPE ?= directory

REPO_ROOT                                      := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))/../..
ifeq ($(VERSION),)
VERSION                                        := $(shell go run ../../api/version/generate/release_generate.go print-rc-version $(CANDIDATE))
endif
COMMIT                                         := $(shell git rev-parse --verify HEAD)
# if EFFECTIVE_VERSION is not set, set it to VERSION+COMMIT
# this is not the same as '?=' because it will also set the value if EFFECTIVE_VERSION is set to an empty string
ifeq ($(EFFECTIVE_VERSION),)
EFFECTIVE_VERSION                              := $(VERSION)+$(COMMIT)
endif
GIT_TREE_STATE                                 := $(shell [ -z "$$(git status --porcelain 2>/dev/null)" ] && echo clean || echo dirty)
PLATFORM                                       := $(shell go env GOOS)/$(shell go env GOARCH)
CACHE_DIR									  := $(shell go env GOCACHE)
MOD_CACHE_DIR								  := $(shell go env GOMODCACHE)


ifneq ($(BUILDX_CACHE_REF),)
	ADDITIONAL_BUILDX_ARGS += --cache-from type=registry,ref=$(BUILDX_CACHE_REF)
	ifeq ($(BUILDX_CACHE_PUSH),true)
		ADDITIONAL_BUILDX_ARGS += --cache-to type=registry,ref=$(BUILDX_CACHE_REF)
	endif
endif

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

$(GEN):
	@mkdir -p $(GEN)

NOW         := $(shell date -u +%FT%T%z)
BUILD_FLAGS := "-s -w \
 -X ocm.software/ocm/api/version.gitVersion=$(EFFECTIVE_VERSION) \
 -X ocm.software/ocm/api/version.gitTreeState=$(GIT_TREE_STATE) \
 -X ocm.software/ocm/api/version.gitCommit=$(COMMIT) \
 -X ocm.software/ocm/api/version.buildDate=$(NOW)"

CMDSRCS=$(shell find $(REPO_ROOT)/cmds/$(NAME) -type f)
OCMSRCS=$(shell find $(REPO_ROOT)/api -type f) $(REPO_ROOT)/go.*

ifeq ($(MULTI),true)
FLAGSUF     = .multi
endif

.PHONY: ctf
ctf: $(GEN)/ctf

$(GEN)/ctf: $(OCM_BIN) $(GEN)/.exists $(GEN)/image.$(NAME)$(FLAGSUF) component-constructor.yaml executorspec.yaml
	@rm -rf "$(GEN)/ctf"
	$(OCM) add componentversions \
		--create \
		--file $(GEN)/ctf \
		--type $(CTF_TYPE) \
		--templater=spiff \
		COMPONENT="$(COMPONENT)" \
		NAME="$(NAME)" \
		VERSION="$(VERSION)" \
		PROVIDER="$(PROVIDER)" \
		COMMIT="$(COMMIT)" \
		GEN="$(GEN)" \
		IMAGE_PLATFORMS="$(IMAGE_PLATFORMS)" \
		MULTI="$(MULTI)" \
		IMAGE="$(IMAGE):$(VERSION)" \
		component-constructor.yaml
	touch "$(GEN)/ctf"

.PHONY: version
version:
	@echo $(VERSION)

.PHONY: build
build: $(GEN)/image.$(NAME)$(FLAGSUF)

$(GEN)/image.$(NAME): $(GEN)/.exists Dockerfile $(CMDSRCS) $(OCMSRCS)
	docker buildx build -t $(IMAGE):$(VERSION) --platform $(PLATFORM) --file Dockerfile $(REPO_ROOT) \
          --build-arg COMMIT=$(COMMIT) \
          --build-arg CACHE_DIR=$(CACHE_DIR) \
          --build-arg MOD_CACHE_DIR=$(MOD_CACHE_DIR) \
          --build-arg EFFECTIVE_VERSION=$(EFFECTIVE_VERSION) \
          --build-arg GIT_TREE_STATE=$(GIT_TREE_STATE) $(ADDITIONAL_BUILDX_ARGS); \
	@touch $(GEN)/image.$(NAME)

push-image:
	docker tag $(IMAGE):$(VERSION) $(OCMREPO)/$(COMPONENT)/$(NAME):$(VERSION)
	docker push $(OCMREPO)/$(COMPONENT)/$(NAME):$(VERSION)

.PHONY: multi
multi: $(GEN)/image.$(NAME).multi

$(GEN)/image.$(NAME).multi: $(GEN)/.exists Dockerfile $(CMDSRCS) $(OCMSRCS)
	for i in $(IMAGE_PLATFORMS); do \
	tag=$$(echo $$i | sed -e s:/:-:g); \
        echo building platform $$i; \
	docker buildx build --load -t $(IMAGE):$(VERSION)-$$tag --platform $$i --file Dockerfile $(REPO_ROOT) \
          --build-arg COMMIT=$(COMMIT) \
          --build-arg CACHE_DIR=$(CACHE_DIR) \
          --build-arg MOD_CACHE_DIR=$(MOD_CACHE_DIR) \
          --build-arg EFFECTIVE_VERSION=$(EFFECTIVE_VERSION) \
          --build-arg GIT_TREE_STATE=$(GIT_TREE_STATE) $(ADDITIONAL_BUILDX_ARGS); \
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
describe: $(OCM_BIN) $(GEN)/ctf
	$(OCM) get resources --lookup $(OCMREPO) -r -o treewide $(GEN)/ctf

.PHONY: descriptor
descriptor: $(OCM_BIN) $(GEN)/ctf
	$(OCM) get component -S v3alpha1 -o yaml $(GEN)/ctf

.PHONY: setup
setup:
	docker run --userns=host --privileged --rm docker/binfmt:a7996909642ee92942dcd6cff44b9b95f08dad64

.PHONY: clean
clean:
	rm -rf $(GEN)
