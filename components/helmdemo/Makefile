NAME         = helmdemo
PROVIDER    ?= ocm.software
GITHUBORG   ?= open-component-model
COMPONENT    = $(PROVIDER)/toi/demo/$(NAME)
OCMREPO     ?= ghcr.io/$(GITHUBORG)/ocm

HELMINSTCOMP = $(PROVIDER)/toi/installers/helminstaller

REPO_ROOT                                     := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))/../..
VERSION                                        = $(shell go run ../../api/version/generate/release_generate.go print-rc-version $(CANDIDATE))
COMMIT                                         = $(shell git rev-parse HEAD)
EFFECTIVE_VERSION                              = $(VERSION)-$(COMMIT)
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

GEN := $(REPO_ROOT)/gen/$(NAME)

CHARTSRCS=$(shell find echoserver -type f)

.PHONY: ctf
ctf: $(GEN)/ctf

$(GEN)/ctf: $(GEN)/ca
	@rm -rf $(GEN)/ctf
	$(OCM) -X keeplocalblob=true transfer ca $(GEN)/ca $(GEN)/ctf
	touch $(GEN)/ctf

.PHONY: version
version:
	@echo $(VERSION)

.PHONY: ca
ca: $(GEN)/ca

$(GEN)/ca: $(GEN)/.exists sources.yaml resources.yaml references.yaml $(CHARTSRCS) packagespec.yaml examples/* helmconfig.yaml
	$(OCM) create ca -f $(COMPONENT) "$(VERSION)" --provider $(PROVIDER) --file $(GEN)/ca
	$(OCM) add sources $(GEN)/ca VERSION="$(VERSION)" COMMIT="$(COMMIT)" sources.yaml
	$(OCM) add resources $(GEN)/ca VERSION="$(VERSION)" COMMIT="$(COMMIT)" resources.yaml
	$(OCM) add references  $(GEN)/ca VERSION="$(VERSION)" COMMIT="$(COMMIT)" HELMINSTCOMP=$(HELMINSTCOMP) HELMINSTVERSION=$(HELMINSTVERSION) references.yaml
	@touch $(GEN)/ca

.PHONY: eval-resources
eval-resources:
	$(OCM) add resources --dry-run VERSION="$(VERSION)" COMMIT="$(COMMIT)" resources.yaml -O "$(GEN)/resources.yaml"

.PHONY: push
push: $(GEN)/ctf $(GEN)/push.$(NAME)

$(GEN)/push.$(NAME): $(GEN)/ctf
	$(OCM) -X keeplocalblob=true transfer ctf -f $(GEN)/ctf $(OCMREPO)
	@touch $(GEN)/push.$(NAME)

.PHONY: plain-push
plain-push: $(GEN)
	$(OCM) -X keeplocalblob=true transfer ctf -f $(GEN)/ctf $(OCMREPO)
	@touch $(GEN)/push.$(NAME)

.PHONY: transport
transport:
ifneq ($(TARGETREPO),)
	$(OCM) -X keeplocalblob=true transfer component -Vc  $(OCMREPO)//$(COMPONENT):$(VERSION) $(TARGETREPO)
endif

$(GEN)/.exists:
	@mkdir -p $(GEN)
	@touch $@

.PHONY: info
info:
	@echo "ROOT:     $(REPO_ROOT)"
	@echo "VERSION:  $(VERSION)"
	@echo "COMMIT:   $(COMMIT)"
	@echo "CREDS:    $(CREDS)"
	@echo "version for helminstaller:  $(HELMINSTVERSION)"

.PHONY: describe
describe: $(GEN)/ctf
	$(OCM) get resources --lookup $(OCMREPO) -r -o treewide $(GEN)/ctf

.PHONY: descriptor
descriptor: $(GEN)/ctf
	$(OCM) get component -S v3alpha1 -o yaml $(GEN)/ctf

.PHONY: clean
clean:
	rm -rf $(GEN)
