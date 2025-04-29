GOPATH                                         := $(shell go env GOPATH)
LOCALBIN := $(shell pwd)/../bin
OS := $(shell go env GOOS 2>/dev/null || sh -c 'uname -o' | sed 's/.*/\L&/' )
ARCH := $(shell go env GOARCH 2>/dev/null || sh -c 'uname -m' | sed 's/.*/\L&/' )
OS_ARCH := $(OS)_$(ARCH)

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

GO_BINDATA_VERSION := "v3.1.3"
GO_BINDATA := $(shell (go-bindata -version 2>/dev/null || echo 0.0.0) | head -n 1 | sed 's/.*go-bindata \([0-9\.]*\).*/\1/')
ifneq ("v$(GO_BINDATA)",$(GO_BINDATA_VERSION))
	deps += go-bindata
endif
VAULT_VERSION := 1.16.2
VAULT := $(shell ($(LOCALBIN)/vault --version 2>/dev/null || echo 0.0) | sed 's/.*Vault v\([0-9\.]*\).*/\1/')
ifneq ($(VAULT), $(VAULT_VERSION))
	deps += vault
endif
OCI_REGISTRY_VERSION := 3.0.0-beta.1
OCI_REGISTRY := $(shell (registry --version 2>/dev/null || echo 0.0) | sed 's/.* v\([0-9a-z\.\-]*\).*/\1/')
ifneq ($(OCI_REGISTRY), $(OCI_REGISTRY_VERSION))
	deps += oci-registry
endif

.PHONY: install-requirements
install-requirements: mkbin $(deps) mdref

mkbin:
	mkdir -p $(LOCALBIN)

.PHONY: go-bindata
go-bindata:
	go install -v  github.com/go-bindata/go-bindata/v3/...@$(GO_BINDATA_VERSION)

.PHONY: vault
vault:
ifneq ($(VAULT), $(VAULT_VERSION))
		curl -o $(LOCALBIN)/vault.zip https://releases.hashicorp.com/vault/$(VAULT_VERSION)/vault_$(VAULT_VERSION)_$(OS_ARCH).zip; \
		unzip -o $(LOCALBIN)/vault.zip -d $(LOCALBIN); \
		rm $(LOCALBIN)/vault.zip; \
		chmod a+x $(LOCALBIN)/vault;
endif

.PHONY: oci-registry
oci-registry:
ifeq (,$(findstring $(OCI_REGISTRY_VERSION), $(OCI_REGISTRY)))
	go install -v github.com/distribution/distribution/v3/cmd/registry@v$(OCI_REGISTRY_VERSION)
endif

.PHONY: mdref
mdref:
	go install -v github.com/mandelsoft/mdref@v0.5.1

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
