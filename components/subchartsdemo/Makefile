NAME             := subcharts
PROVIDER         ?= ocm.software
GITHUBORG        ?= open-component-model
COMPONENT_PREFIX  = $(PROVIDER)/toi/demo/$(NAME)
OCMREPO          ?= ghcr.io/$(GITHUBORG)/ocm

HELMINSTCOMP      = ocm.software/toi/installers/helminstaller

ECHO_VERSION = "1.10"
ECHO_CHART_VERSION = 0.1.0
PODINFO_VERSION = 6.3.5
PODINFO_CHART_VERSION = 6.3.5

REPO_ROOT                                     := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))/../..
GIT_TREE_STATE                                 = $(shell [ -z "$$(git status --porcelain 2>/dev/null)" ] && echo clean || echo dirty)
VERSION                                        = $(shell go run $(REPO_ROOT)/api/version/generate/release_generate.go print-rc-version $(CANDIDATE))
COMMIT                                         = $(shell git rev-parse --verify HEAD)
# if EFFECTIVE_VERSION is not set, set it to VERSION-COMMIT
# this is not the same as '?=' because it will also set the value if EFFECTIVE_VERSION is set to an empty string
ifeq ($(EFFECTIVE_VERSION),)
EFFECTIVE_VERSION                              := $(VERSION)+$(COMMIT)
endif
HELMINSTVERSION                               ?= $(VERSION)

CREDS ?=
# Define the path to the binary
OCM_BIN = $(REPO_ROOT)/bin/ocm

# Rule to build the binary if it doesn't exist or if the source code has changed
$(OCM_BIN): $(REPO_ROOT)/cmds/ocm/main.go
	mkdir -p $(REPO_ROOT)/bin
	go build -ldflags $(BUILD_FLAGS) -o $(OCM_BIN) $(REPO_ROOT)/cmds/ocm

# Use the binary for the OCM command
OCM = $(OCM_BIN) $(CREDS)

NOW         := $(shell date -u +%FT%T%z)
BUILD_FLAGS := "-s -w \
 -X ocm.software/ocm/api/version.gitVersion=$(EFFECTIVE_VERSION) \
 -X ocm.software/ocm/api/version.gitTreeState=$(GIT_TREE_STATE) \
 -X ocm.software/ocm/api/version.gitCommit=$(COMMIT) \
 -X ocm.software/ocm/api/version.buildDate=$(NOW)"

GEN = $(REPO_ROOT)/gen/subchartsdemo

ECHOCHARTSRCS=$(shell find echoserver -type f)

.PHONY: ctf
ctf: $(GEN)/ctf

.PHONY: version
version:
	@echo $(VERSION)

$(GEN)/ctf: $(GEN)/.exists component-constructor.yaml $(ECHOCHARTSRCS) packagespec.yaml podinfo/podinfo-$(PODINFO_CHART_VERSION).tgz $(OCM_BIN)
	@rm -rf $(GEN)/ctf
	$(OCM) add componentversions --create VERSION="$(VERSION)" COMMIT="$(COMMIT)" COMPONENT_PREFIX=$(COMPONENT_PREFIX) PROVIDER=$(PROVIDER) PODINFO_VERSION=$(PODINFO_VERSION) PODINFO_CHART_VERSION=$(PODINFO_CHART_VERSION) HELMINSTCOMP=$(HELMINSTCOMP) HELMINSTVERSION=$(HELMINSTVERSION) ECHO_VERSION=$(ECHO_VERSION) ECHO_CHART_VERSION=$(ECHO_CHART_VERSION) --file $(GEN)/ctf component-constructor.yaml
	@touch $(GEN)/ctf

.PHONY: push
push: $(GEN)/ctf $(GEN)/push.$(NAME)

$(GEN)/push.$(NAME): $(GEN)/ctf $(OCM_BIN)
	$(OCM) -X keeplocalblob=true transfer ctf --copy-resources -f $(GEN)/ctf $(OCMREPO)
	@touch $(GEN)/push.$(NAME)

.PHONY: plain-push
plain-push: $(GEN) $(OCM_BIN)
	$(OCM) -X keeplocalblob=true transfer ctf --copy-resources -f $(GEN)/ctf $(OCMREPO)
	@touch $(GEN)/push.$(NAME)

$(GEN)/.exists:
	@mkdir -p $(GEN)
	@touch $@

.PHONY: info
info:
	@echo "ROOT:     $(REPO_ROOT)"
	@echo "VERSION:  $(VERSION)"
	@echo "COMMIT:   $(COMMIT)"
	@echo "CREDS:    $(CREDS)"

.PHONY: describe
describe: $(GEN)/ctf $(OCM_BIN)
	$(OCM) get resources --lookup $(OCMREPO) -r -o treewide $(GEN)/ctf

.PHONY: descriptor
descriptor: $(GEN)/ctf $(OCM_BIN)
	$(OCM) get component -S v3alpha1 -o yaml $(GEN)/ctf

.PHONY: clean
clean:
	rm -rf $(GEN)
