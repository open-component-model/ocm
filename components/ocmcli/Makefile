NAME      = ocmcli
CMD       = ocm
PROVIDER  ?= ocm.software
IMAGE     := $(NAME)-image
GITHUBORG  ?= open-component-model
COMPONENT = $(PROVIDER)/$(NAME)
OCMREPO   ?= ghcr.io/$(GITHUBORG)/ocm
MULTI     ?= true
IMAGE_PLATFORMS ?= linux/amd64 linux/arm64
PLATFORMS ?= $(IMAGE_PLATFORMS) darwin/arm64 darwin/amd64 windows/amd64
CTF_TYPE ?= directory

REPO_ROOT                                      := $(dir $(realpath $(lastword $(MAKEFILE_LIST))))../..
GIT_TREE_STATE                                 = $(shell [ -z "$$(git status --porcelain 2>/dev/null)" ] && echo clean || echo dirty)
ifeq ($(VERSION),)
VERSION                                        := $(shell go run ../../api/version/generate/release_generate.go print-rc-version $(CANDIDATE))
endif
COMMIT                                         = $(shell git rev-parse --verify HEAD)
# if EFFECTIVE_VERSION is not set, set it to VERSION+COMMIT
# this is not the same as '?=' because it will also set the value if EFFECTIVE_VERSION is set to an empty string
ifeq ($(EFFECTIVE_VERSION),)
EFFECTIVE_VERSION                              := $(VERSION)+$(COMMIT)
endif
PLATFORM_OS                                    := $(shell go env GOOS)
PLATFORM_ARCH                                  := $(shell go env GOARCH)

CMDSRCS=$(shell find $(REPO_ROOT)/cmds/$(CMD) -type f) Makefile
OCMSRCS=$(shell find $(REPO_ROOT)/api -type f) $(REPO_ROOT)/go.*

ifeq ($(MULTI),true)
FLAGSUF     = .multi
endif

ifneq ($(BUILDX_CACHE_REF),)
	ADDITIONAL_BUILDX_ARGS += --cache-from type=registry,ref=$(BUILDX_CACHE_REF)
	ifeq ($(BUILDX_CACHE_PUSH),true)
		ADDITIONAL_BUILDX_ARGS += --cache-to type=registry,ref=$(BUILDX_CACHE_REF),mode=max
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

.PHONY: build
build: $(GEN)/build

$(GEN)/build: $(GEN) $(GEN)/.exists $(CMDSRCS) $(OCMSRCS)
	@for i in $(PLATFORMS); do \
		tag=$$(echo $$i | sed -e s:/:-:g); \
		echo GOARCH=$$(basename $$i) GOOS=$$(dirname $$i) CGO_ENABLED=0 go build -ldflags $(BUILD_FLAGS) -o $(GEN)/$(NAME).$$tag ../../cmds/$(CMD); \
		GOARCH=$$(basename $$i) GOOS=$$(dirname $$i) CGO_ENABLED=0 go build -ldflags $(BUILD_FLAGS) -o $(GEN)/$(NAME).$$tag ../../cmds/$(CMD) & \
	done; \
	wait
	@touch $(GEN)/build

.PHONY: image
image:
	@PLATFORM_OS_OVERRIDE=$(PLATFORM_OS); \
	if [ $$PLATFORM_OS_OVERRIDE == darwin ]; then \
		echo; echo "Building linux instead of darwin as there's no native Docker platform for darwin"; echo; \
		docker buildx build -t $(IMAGE):$(VERSION) --platform linux/$(PLATFORM_ARCH) --file $(REPO_ROOT)/Dockerfile $(REPO_ROOT) \
			--label org.opencontainers.image.version=$(VERSION) \
			--label org.opencontainers.image.revision=$(VERSION) $(ADDITIONAL_BUILDX_ARGS); \
	else \
		echo; echo "Building for $(PLATFORM_OS)/$(ARCH)"; echo; \
		docker buildx build -t $(IMAGE):$(VERSION) --platform $(PLATFORM_OS)/$(PLATFORM_ARCH) --file $(REPO_ROOT)/Dockerfile $(REPO_ROOT) \
			--label org.opencontainers.image.version=$(VERSION) \
			--label org.opencontainers.image.revision=$(VERSION) $(ADDITIONAL_BUILDX_ARGS); \
	fi

.PHONY: image.multi
image.multi:
	for i in $(IMAGE_PLATFORMS); do \
		tag=$$(echo $$i | sed -e s:/:-:g); \
			echo building platform $$i; \
		docker buildx build --load -t $(IMAGE):$(VERSION)-$$tag --platform $$i --file $(REPO_ROOT)/Dockerfile $(REPO_ROOT) \
			--build-arg OCM_VERSION=$(EFFECTIVE_VERSION) \
			--label org.opencontainers.image.version=$(VERSION) \
			--label org.opencontainers.image.revision=$(VERSION) $(ADDITIONAL_BUILDX_ARGS); \
	done

.PHONY: ctf
ctf: $(GEN)/ctf

$(GEN)/ctf: $(OCM_BIN) $(GEN)/.exists $(GEN)/build image$(FLAGSUF) component-constructor.yaml $(CHARTSRCS) Makefile
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
		PLATFORMS="$(PLATFORMS)" \
		IMAGE_PLATFORMS="$(IMAGE_PLATFORMS)" \
		MULTI=$(MULTI) \
		IMAGE="$(IMAGE):$(VERSION)" \
		component-constructor.yaml
	touch "$(GEN)/ctf"

.PHONY: version
version:
	@echo $(VERSION)

.PHONY: push
push: $(GEN)/ctf $(GEN)/push.$(NAME)

$(GEN)/push.$(NAME): $(GEN)/ctf $(OCM_BIN)
	$(OCM) transfer ctf -f $(GEN)/ctf $(OCMREPO)
	@touch $(GEN)/push.$(NAME)

.PHONY: plain-push
plain-push: $(GEN) $(OCM_BIN)
	$(OCM) transfer ctf -f $(GEN)/ctf $(OCMREPO)
	@touch $(GEN)/push.$(NAME)

.PHONY: transport
transport: $(OCM_BIN)
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
describe: $(GEN)/ctf $(OCM_BIN)
	$(OCM) get resources --lookup $(OCMREPO) -r -o treewide $(GEN)/ctf

.PHONY: descriptor
descriptor: $(GEN)/ctf $(OCM_BIN)
	$(OCM) get component -S v3alpha1 -o yaml $(GEN)/ctf

.PHONY: clean
clean:
	rm -rf $(GEN)
