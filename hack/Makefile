# SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
#
# SPDX-License-Identifier: Apache-2.0

GOPATH                                         := $(shell go env GOPATH)

ifeq ($(OS),Windows_NT)
	detected_OS := Windows
else
	detected_OS := $(shell sh -c 'uname 2>/dev/null || echo Unknown')
endif

deps := 
# TODO exact versions to compare
gSED := $(shell (sed --version 2>/dev/null || echo 0.0) | head -n 1 | sed 's/.*(GNU sed) \([0-9\.]*\).*/\1/')
ifeq ("v$(gSED)","v0.0")
	deps += $(detected_OS)_sed
endif
gTAR := $(shell (tar --version 2>/dev/null || echo 0.0) | head -n 1 | sed 's/.*(GNU tar) \([0-9\.]*\).*/\1/')
ifeq ("v$(gTAR)","v0.0")
	deps += $(detected_OS)_tar
endif
gCOREUTILS := $(shell (basename --version 2>/dev/null || echo 0.0) | head -n 1 | sed 's/.*(GNU coreutils) \([0-9\.]*\).*/\1/')
ifeq ("v$(gCOREUTILS)","v0.0")
	deps += $(detected_OS)_coreutils
endif
gGREP := $(shell (grep --version 2>/dev/null || echo 0.0) | head -n 1 | sed 's/.*(GNU grep) \([0-9\.]*\).*/\1/')
ifeq ("v$(gGREP)","v0.0")
	deps += $(detected_OS)_grep
endif
JQ := $(shell (jq --version 2>/dev/null || echo 0.0) | sed 's/.*-\([0-9\.]*\).*/\1/')
ifeq ("v$(JQ)","v0.0")
	deps += $(detected_OS)_jq
endif

GOLANGCILINT_VERSION := "v1.49.0"
GOLANGCILINTV := $(shell (golangci-lint --version 2>/dev/null || echo 0.0.0) | sed 's/.*v\([0-9\.]*\) .*/\1/')
ifneq ("v$(GOLANGCILINTV)",$(GOLANGCILINT_VERSION))
  deps += golangci-lint
endif
GO_BINDATA_VERSION := "v3.1.3"
GO_BINDATA := $(shell (go-bindata -version 2>/dev/null || echo 0.0.0) | head -n 1 | sed 's/.*go-bindata \([0-9\.]*\).*/\1/')
ifneq ("v$(GO_BINDATA)",$(GO_BINDATA_VERSION))
	deps += go-bindata
endif

.PHONY: install-requirements
install-requirements: $(deps) $(GOPATH)/bin/goimports
#	@$(REPO_ROOT)/hack/install-requirements.sh

.PHONY: golangci-lint
golangci-lint:
	go install -v github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCILINT_VERSION)

.PHONY: go-bindata
go-bindata:
	go install -v  github.com/go-bindata/go-bindata/v3/...@$(GO_BINDATA_VERSION)

$(GOPATH)/bin/goimports:
	go install -v golang.org/x/tools/cmd/goimports@latest

Linux_jq:
	$(info -> jq is missing)
	$(info -  sudo apt-get install jq / sudo dnf install jq / sudo zypper install jq / sudo pacman -S jq)

Darwin_sed: Darwin
	$(info -> GNU sed is missing)
	$(info -  brew install gnu-sed)
	$(info -  export PATH=/usr/local/opt/gnu-sed/libexec/gnubin:$$PATH)

Darwin_tar: Darwin
	$(info -> GNU tar is missing)
	$(info -  brew install gnu-tar)
	$(info -  export PATH=/usr/local/opt/gnu-tar/libexec/gnubin:$$PATH)

Darwin_grep: Darwin
	$(info -> GNU grep is missing)
	$(info -  brew install grep)
	$(info -  export PATH=/usr/local/opt/grep/libexec/gnubin:$$PATH)

Darwin_coreutils: Darwin
	$(info -> GNU Core Utils are missing)
	$(info -  brew install coreutils)
	$(info -  export PATH=/usr/local/opt/coreutils/libexec/gnubin:$$PATH)

Darwin_jq: Darwin
	$(info -> jq is missing)
	$(info -  brew install jq)

.PHONY: Darwin
Darwin:
	$(info You are running in a MAC OS environment!)
	$(info Please make sure you have installed the following tools.)
	$(info Please allow all GNU tools to be used without their "g" prefix.)
