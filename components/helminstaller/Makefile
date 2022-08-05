NAME      = helminstaller
IMAGE     = mandelsoft/ocm$(NAME)
COMPONENT = github.com/$(IMAGE)


REPO_ROOT                                      := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))/../..
VERSION                                        = $(shell git describe --tags --exact-match 2>/dev/null|| echo "$$(cat $(REPO_ROOT)/VERSION)-dev")
COMMIT                                         = $(shell git rev-parse HEAD)
EFFECTIVE_VERSION                              = $(VERSION)-$(COMMIT)

GEN = $(REPO_ROOT)/gen/$(NAME)
OCM = go run $(REPO_ROOT)/cmds/ocm

CMDSRCS=$(shell find $(REPO_ROOT)/cmds/$(NAME) -type f)
OCMSRCS=$(shell find $(REPO_ROOT)/pkg -type f) $(REPO_ROOT)/go.*

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

$(GEN)/ca: $(GEN)/image.$(NAME) $(GEN) resources.yaml
	$(OCM) create ca -f $(COMPONENT) "$(VERSION)" mandelsoft $(GEN)/ca
	$(OCM) add resources $(GEN)/ca VERSION="$(VERSION)" COMMIT="$(COMMIT)" IMAGE="$(IMAGE):$(VERSION)" resources.yaml
	touch $(GEN)/ca


.PHONY: build
build: $(GEN)/image.$(NAME)

$(GEN)/image.$(NAME): Dockerfile $(CMDSRCS) $(OCMSRCS)
	docker build -t $(IMAGE):$(VERSION) --file Dockerfile $(REPO_ROOT)
	touch $(GEN)/image.$(NAME)

.PHONY: push
push: $(GEN)/ctf $(GEN)/push.$(NAME)

$(GEN)/push.$(NAME): $(GEN)/ctf
	$(OCM) transfer ctf -f $(GEN)/ctf ghcr.io/mandelsoft/cnudie
	touch $(GEN)/push.$(NAME)

$(GEN):
	mkdir -p $(GEN)

.PHONY: info
info:
	@echo "ROOT:     $(REPO_ROOT)"
	@echo "VERSION:  $(VERSION)"
	@echo "COMMIT;   $(COMMIT)"

.PHONY: clean
clean:
	rm -rf $(GEN)
