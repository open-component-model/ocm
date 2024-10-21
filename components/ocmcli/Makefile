NAME      = ocmcli
CMD       = ocm
PROVIDER  ?= ocm.software
IMAGE     := $(NAME)-image
GITHUBORG  ?= open-component-model
COMPONENT = $(PROVIDER)/$(NAME)
OCMREPO   ?= ghcr.io/$(GITHUBORG)/ocm
MULTI     ?= true
IMAGE_PLATFORMS ?= linux/amd64 linux/arm64
PLATFORMS = $(IMAGE_PLATFORMS) darwin/arm64 darwin/amd64 windows/amd64

REPO_ROOT                                      := $(dir $(realpath $(lastword $(MAKEFILE_LIST))))../..
GIT_TREE_STATE                                 = $(shell [ -z "$$(git status --porcelain 2>/dev/null)" ] && echo clean || echo dirty)
VERSION                                        = $(shell go run ../../api/version/generate/release_generate.go print-rc-version $(CANDIDATE))
COMMIT                                         = $(shell git rev-parse HEAD)
EFFECTIVE_VERSION                              = $(VERSION)+$(COMMIT)
PLATFORM_OS                                    := $(shell go env GOOS)
PLATFORM_ARCH                                  := $(shell go env GOARCH)

CMDSRCS=$(shell find $(REPO_ROOT)/cmds/$(CMD) -type f) Makefile
OCMSRCS=$(shell find $(REPO_ROOT)/pkg -type f) $(REPO_ROOT)/go.*

ATTRIBUTES = VERSION="$(VERSION)" NAME="$(NAME)" COMMIT="$(COMMIT)" IMAGE="$(IMAGE):$(VERSION)" PLATFORMS="$(PLATFORMS)" IMAGE_PLATFORMS="$(IMAGE_PLATFORMS)" GEN="$(GEN)" MULTI=$(MULTI)

ifeq ($(MULTI),true)
FLAGSUF     = .multi
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

GEN = $(REPO_ROOT)/gen/$(shell basename $(realpath .))

NOW         := $(shell date -u +%FT%T%z)
BUILD_FLAGS := "-s -w \
 -X ocm.software/ocm/api/version.gitVersion=$(EFFECTIVE_VERSION) \
 -X ocm.software/ocm/api/version.gitTreeState=$(GIT_TREE_STATE) \
 -X ocm.software/ocm/api/version.gitCommit=$(COMMIT) \
 -X ocm.software/ocm/api/version.buildDate=$(NOW)"

ALPINE_LATEST_VER=$(shell curl -s https://registry.hub.docker.com/v2/repositories/library/alpine/tags | jq '.results[1].name' | xargs)

.PHONY: build
build: $(GEN)/build

$(GEN)/build: $(GEN)/.exists $(CMDSRCS) $(OCMSRCS)
	@for i in $(PLATFORMS); do \
		tag=$$(echo $$i | sed -e s:/:-:g); \
		echo GOARCH=$$(basename $$i) GOOS=$$(dirname $$i) CGO_ENABLED=0 go build -ldflags $(BUILD_FLAGS) -o $(GEN)/$(NAME).$$tag ../../cmds/$(CMD); \
		GOARCH=$$(basename $$i) GOOS=$$(dirname $$i) CGO_ENABLED=0 go build -ldflags $(BUILD_FLAGS) -o $(GEN)/$(NAME).$$tag ../../cmds/$(CMD) & \
	done; \
	wait
	@touch $(GEN)/build

.PHONY: image
image: $(GEN)/image

$(GEN)/image: $(GEN)/.exists Dockerfile $(GEN)/build
	@PLATFORM_OS_OVERRIDE=$(PLATFORM_OS); \
	if [ $$PLATFORM_OS_OVERRIDE == darwin ]; then \
		echo; echo "Building linux instead of darwin as there's no native Docker platform for darwin"; echo; \
		docker buildx build -t $(IMAGE):$(VERSION) --platform linux/$(PLATFORM_ARCH) --file Dockerfile $(REPO_ROOT) \
			--build-arg OCM_VERSION=$(EFFECTIVE_VERSION) \
			--build-arg ALPINE_VERSION=$(ALPINE_LATEST_VER) \
			--build-arg SUFFIX=$$(echo linux/$(PLATFORM_ARCH) | sed -e s:/:-:g); \
	else \
		echo; echo "Building for $(PLATFORM_OS)/$(ARCH)"; echo; \
		docker buildx build -t $(IMAGE):$(VERSION) --platform $(PLATFORM_OS)/$(PLATFORM_ARCH) --file Dockerfile $(REPO_ROOT) \
			--build-arg OCM_VERSION=$(EFFECTIVE_VERSION) \
			--build-arg ALPINE_VERSION=$(ALPINE_LATEST_VER) \
			--build-arg SUFFIX=$$(echo $(PLATFORM_OS)/$(PLATFORM_ARCH) | sed -e s:/:-:g); \
	fi
	@touch $(GEN)/image

.PHONY: image.multi
image.multi: $(GEN)/image.multi

$(GEN)/image.multi: Dockerfile $(GEN)/build
	echo "Building with Alpine version: ${ALPINE_LATEST_VER}"
	for i in $(IMAGE_PLATFORMS); do \
	tag=$$(echo $$i | sed -e s:/:-:g); \
        echo building platform $$i; \
	docker buildx build --load -t $(IMAGE):$(VERSION)-$$tag --platform $$i --file Dockerfile $(REPO_ROOT) \
          --build-arg OCM_VERSION=$(EFFECTIVE_VERSION) \
		  --build-arg ALPINE_VERSION=$(ALPINE_LATEST_VER) \
          --build-arg SUFFIX=$$tag; \
	done
	@touch $(GEN)/image.multi

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

$(GEN)/ca.done: $(GEN)/.exists $(GEN)/build $(GEN)/image$(FLAGSUF) resources.yaml $(CHARTSRCS) Makefile
	$(OCM) create ca -f $(COMPONENT) "$(VERSION)" --provider $(PROVIDER) --file $(GEN)/ca
	$(OCM) add resources --templater=spiff --file $(GEN)/ca $(ATTRIBUTES) resources.yaml
	$(OCM) add sources $(GEN)/ca VERSION="$(VERSION)" COMMIT="$(COMMIT)" sources.yaml
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
