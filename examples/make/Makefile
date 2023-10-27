COMPONENT := github.com/mandelsoft/playground
OCI_REPO := ghcr.io/mandelsoft/cnudie

REPO_ROOT                                      := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
VERSION                                        = $(shell cat $(REPO_ROOT)/VERSION)
COMMIT                                         = $(shell git rev-parse HEAD)
EFFECTIVE_VERSION                              = $(VERSION)-$(COMMIT)

.PHONY: ctf
ctf: ca ## Create CTF from component archive
	ocm transfer ca gen/ca gen/ctf
	
.PHONY: ca
ca: gen releasenotes ## Create component archive
	ocm create ca -f $(COMPONENT) "$(VERSION)" mandelsoft gen/ca
	ocm add resources --file gen/ca VERSION="$(VERSION)" COMMIT="$(COMMIT)" resources.yaml

.PHONY: dummy
dummy: commit ca ## Commit and create component archive

.PHONY: patch
patch: clean incpatch ctf ## Create a new patch version with CTF

.PHONY: minor
minor: clean incminor ctf ## Create a new minor version with CTF

.PHONY: major
major: clean incmajor ctf ## Create a new major version with CTF


.PHONY: release-patch
release-patch: clean incpatch release ## Release the component with a new patch version

.PHONY: release-minor
release-minor: clean incminor release ## Release the component with a new minor version

.PHONY: release-major
release-major: clean incmajor release ## Release the component with a new major version


.PHONY: incpatch
incpatch: ## Increment patch version of the component
	semver -i patch $(VERSION) | tee VERSION

.PHONY: incminor
incminor: ## Increment minor version of the component
	semver -i minor $(VERSION) | tee VERSION

.PHONY: incmajor
incmajor: ## Increment major version of the component
	semver -i major $(VERSION) | tee VERSION

.PHONY: releasenotes
releasenotes: content/RELEASENOTES.md ## Generate release notes
content/RELEASENOTES.md: VERSION
	cp content/RELEASENOTES.md /tmp
	echo '- **Release '"`cat VERSION`"'**\n' > content/RELEASENOTES.md
	cat /tmp/RELEASENOTES.md >>content/RELEASENOTES.md

.PHONY: push 
push: ctf ## Push local CTF file into a OCI repository
	ocm transfer ctf -f gen/ctf $(OCI_REPO)

.PHONY: gen
gen: ## Create "gen" directory
	mkdir -p gen

.PHONY: commit
commit: releasenotes ## Commit changes
	git add .
	git commit -m "release $(VERSION)"

.PHONY: release
release: commit push ## Release component
	
.PHONY: info
info: ## Display information about the component
	@echo "VERSION:  $(VERSION)"
	@echo "COMMIT:   $(COMMIT)"

.PHONY: clean
clean: ## Delete generated content
	rm -rf gen

.PHONY: help
help:  ## Display this help. Thanks to https://www.thapaliya.com/en/writings/well-documented-makefiles/
	@echo "Targets to create and release the component with new version."
ifeq ($(OS),Windows_NT)
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make <target>\n"} /^[0-9a-zA-Z_-]+:.*?##/ { printf "  %-40s %s\n", $$1, $$2 } /^##@/ { printf "\n%s\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
else
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[0-9a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-40s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
endif
