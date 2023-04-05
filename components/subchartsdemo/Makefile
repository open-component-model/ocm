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
VERSION                                        = $(shell go run $(REPO_ROOT)/pkg/version/generate/release_generate.go print-rc-version $(CANDIDATE))
COMMIT                                         = $(shell git rev-parse HEAD)
EFFECTIVE_VERSION                              = $(VERSION)-$(COMMIT)
HELMINSTVERSION                               ?= $(VERSION)
CREDS                                          = $(shell $(REPO_ROOT)/hash/githubcreds.sh)

CREDS := $(shell $(REPO_ROOT)/hack/githubcreds.sh)
OCM = go run $(REPO_ROOT)/cmds/ocm $(CREDS)

GEN = $(REPO_ROOT)/gen/subchartsdemo

ECHOCHARTSRCS=$(shell find echoserver -type f)

.PHONY: ctf
ctf: $(GEN)/ctf

.PHONY: version
version:
	@echo $(VERSION)

$(GEN)/ctf: $(GEN)/.exists components.yaml $(ECHOCHARTSRCS) packagespec.yaml podinfo/podinfo-$(PODINFO_CHART_VERSION).tgz
	@rm -rf $(GEN)/ctf
	$(OCM) add componentversions --create VERSION="$(VERSION)" COMMIT="$(COMMIT)" COMPONENT_PREFIX=$(COMPONENT_PREFIX) PROVIDER=$(PROVIDER) PODINFO_VERSION=$(PODINFO_VERSION) PODINFO_CHART_VERSION=$(PODINFO_CHART_VERSION) HELMINSTCOMP=$(HELMINSTCOMP) HELMINSTVERSION=$(HELMINSTVERSION) ECHO_VERSION=$(ECHO_VERSION) ECHO_CHART_VERSION=$(ECHO_CHART_VERSION) --file $(GEN)/ctf components.yaml
	@touch $(GEN)/ctf

.PHONY: push
push: $(GEN)/ctf $(GEN)/push.$(NAME)

$(GEN)/push.$(NAME): $(GEN)/ctf
	$(OCM) -X keeplocalblob=true transfer ctf --copy-resources -f $(GEN)/ctf $(OCMREPO)
	@touch $(GEN)/push.$(NAME)

.PHONY: plain-push
plain-push: $(GEN)
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
describe: $(GEN)/ctf
	ocm get resources --lookup $(OCMREPO) -o treewide $(GEN)/ctf

.PHONY: descriptor
descriptor: $(GEN)/ctf
	ocm get component -S v3alpha1 -o yaml $(GEN)/ctf

.PHONY: clean
clean:
	rm -rf $(GEN)
