NAME      = helmdemo
PROVIDER  ?= mandelsoft
IMAGE     = ocm$(NAME)
COMPONENT = github.com/$(PROVIDER)/ocm$(NAME)
OCMREPO   ?= ghcr.io/$(PROVIDER)/cnudie

HELMINSTCOMP = github.com/$(PROVIDER)/ocmhelminstaller

REPO_ROOT                                      := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))/../..
VERSION                                        = $(shell git describe --tags --exact-match 2>/dev/null|| echo "$$(cat $(REPO_ROOT)/VERSION)-dev")
COMMIT                                         = $(shell git rev-parse HEAD)
EFFECTIVE_VERSION                              = $(VERSION)-$(COMMIT)

GEN = $(REPO_ROOT)/gen/$(NAME)
OCM = go run $(REPO_ROOT)/cmds/ocm

CHARTSRCS=$(shell find echoserver -type f)

.PHONY: ctf
ctf: $(GEN)/ctf

$(GEN)/ctf: $(GEN)/ca
	$(OCM) transfer ca $(GEN)/ca $(GEN)/ctf
	touch $(GEN)/ctf

.PHONY: version
version:
	@echo $(VERSION)

.PHONY: ca
ca: $(GEN)/ca

$(GEN)/ca: $(GEN)/.exists sources.yaml resources.yaml references.yaml $(CHARTSRCS) packagespec.yaml helmconfig.yaml
	$(OCM) create ca -f $(COMPONENT) "$(VERSION)" $(PROVIDER) $(GEN)/ca
	$(OCM) add sources $(GEN)/ca VERSION="$(VERSION)" COMMIT="$(COMMIT)" IMAGE="$(IMAGE):$(VERSION)" sources.yaml
	$(OCM) add resources $(GEN)/ca VERSION="$(VERSION)" COMMIT="$(COMMIT)" IMAGE="$(IMAGE):$(VERSION)" resources.yaml
	$(OCM) add references $(GEN)/ca VERSION="$(VERSION)" COMMIT="$(COMMIT)" IMAGE="$(IMAGE):$(VERSION)" HELMINSTCOMP=$(HELMINSTCOMP) references.yaml
	@touch $(GEN)/ca

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
	@echo "COMMIT;   $(COMMIT)"

.PHONY: describe
describe: $(GEN)/ctf
	ocm get resources --lookup $(OCMREPO) -c -o treewide $(GEN)/ctf

.PHONY: descriptor
descriptor: $(GEN)/ctf
	ocm get component -S v3alpha1 -o yaml $(GEN)/ctf

.PHONY: clean
clean:
	rm -rf $(GEN)
