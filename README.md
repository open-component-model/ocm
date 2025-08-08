# Open Component Model

CHANGE

[![OpenSSF Best Practices](https://bestpractices.coreinfrastructure.org/projects/7156/badge)](https://bestpractices.coreinfrastructure.org/projects/7156)
[![REUSE status](https://api.reuse.software/badge/github.com/open-component-model/ocm)](https://api.reuse.software/info/github.com/open-component-model/ocm)
[![OCM Integration Tests](https://github.com/open-component-model/ocm-integrationtest/actions/workflows/integrationtest.yaml/badge.svg?branch=main)](https://open-component-model.github.io/ocm-integrationtest/report.html)
[![Go Report Card](https://goreportcard.com/badge/ocm.software/ocm)](https://goreportcard.com/report/ocm.software/ocm)

The Open Component Model (OCM) is an open standard to describe software bills of delivery (SBOD). OCM is a technology-agnostic and machine-readable format focused on the software artifacts that must be delivered for software products.

Check out the [the main OCM project web page](https://ocm.software) to find out what OCM offers you for implementing a secure software supply chain. It is your central entry point to all kind of OCM related [docs and guides](https://ocm.software/docs/overview/about), the [OCM specification](https://ocm.software/docs/overview/specification/) and all project [github repositories](https://github.com/open-component-model). It also offers a [Getting Started](https://ocm.software/docs/getting-started/) to quickly make your hands dirty with OCM, its toolset and concepts :smiley:

## OCM Specifications

OCM describes delivery [artifacts](https://github.com/open-component-model/ocm-spec/tree/main/doc/01-model/02-elements-toplevel.md#artifacts-resources-and-sources) that can be accessed from many types of [component repositories](https://github.com/open-component-model/ocm-spec/tree/main/doc/01-model/01-model.md#component-repositories). It defines a set of semantic, formatting, and other types of specifications that can be found in the [`ocm-spec` repository](https://github.com/open-component-model/ocm-spec). Start learning about the core concepts of OCM elements [here](https://github.com/open-component-model/ocm-spec/tree/main/doc/01-model/02-elements-toplevel.md#model-elements).

## OCM Library

This project provides a Go library containing an API for interacting with the
[Open Component Model (OCM)](https://github.com/open-component-model/ocm-spec) elements and mechanisms.

The library currently supports the following [repository mappings](https://github.com/open-component-model/ocm-spec/tree/main/doc/03-persistence/02-mappings.md#mappings-for-ocm-persistence):

- **OCI**: Use the repository prefix path of an OCI repository to implement an OCM
  repository.
- **CTF (Common Transport Format)**: Use a file-based binding to represent any set of
  component versions as filesystem content (directory, tar, tgz).
- **Component Archive**: Compose the content of a component version on the
  filesystem.

For the usage of the library to access OCM repositories, handle configuration and credentials see the [examples section](examples/lib/README.md).

Additionally, OCM provides a generic solution for how to:

- Sign component versions in any supported OCM repository implementation.
- Verify signatures based on public keys or verified certificates.
- Transport component versions, per reference or as values to any of the
  repository implementations.

## [OCM CLI](docs/reference/ocm.md)

The [`ocm` CLI](docs/reference/ocm.md) may also be used to interact with OCM mechanisms. It makes it easy to create component versions and embed them in build processes.

The code for the CLI can be found in [package `cmds/ocm`](cmds/ocm).

The OCI and OCM support can be found in packages
[`api/oci`](api/oci) and [`api/ocm`](api/ocm).

## Installation

Install the latest release with

- [Bash](#bash)
- [Homebrew](#homebrew)
- [NixOS](#nixos)
- [AUR](#aur)
- [Docker and Podman](#container)
- [Chocolatey](#chocolatey)

### Bash

To install with `bash` for macOS or Linux execute the following command:

```bash
curl -s https://ocm.software/install.sh | sudo bash
```

### Homebrew

Install using [Homebrew](https://brew.sh)

```bash
# Homebrew (macOS and Linux)
brew install open-component-model/tap/ocm
```

### NixOS

Install using [Nix](https://nixos.org) (with [Flakes](https://nixos.wiki/wiki/Flakes))

```bash
# Nix (macOS, Linux, and Windows)
# ad-hoc cmd execution
nix run github:open-component-model/ocm -- --help
nix run github:open-component-model/ocm#helminstaller -- --help

# install development version
nix profile install github:open-component-model/ocm
# or release <version>
nix profile install github:open-component-model/ocm/<version>

#check installation
nix profile list | grep ocm

# optionally, open a new shell and verify that cmd completion works
ocm --help
```

### AUR

Install from [AUR (Arch Linux User Repository)](https://archlinux.org/)

package-url: [aur.archlinux.org/packages/ocm-cli](https://aur.archlinux.org/packages/ocm-cli)

```bash
# if not using a helper util
git clone https://aur.archlinux.org/ocm-cli.git
cd ocm-cli
makepkg -i
```

[AUR Documentation](https://wiki.archlinux.org/title/Arch_User_Repository)

### Container

Usage via [Docker](https://www.docker.com/) / [Podman](https://podman.io/)

```bash
docker run -t ghcr.io/open-component-model/ocm:latest --help
```

```bash
podman run -t ghcr.io/open-component-model/ocm:latest --help
```

#### Build and run it yourself

```bash
podman build -t ocm .
podman run --rm -t ocm --loglevel debug --help
```

or interactively:

```bash
podman run --rm -it ocm /bin/sh
```

You can pass in the following arguments to override the predefined defaults:

- `GO_VERSION`: The **golang** version to be used for compiling.
- `ALPINE_VERSION`: The **alpine** version to be used as the base image.
- `GO_PROXY`: Your **go** proxy to be used for fetching dependencies.

Please check [hub.docker.com](https://hub.docker.com/_/golang/tags?page=1&name=alpine) for possible version combinations.

```bash
podman build -t ocm --build-arg GO_VERSION=1.24 --build-arg ALPINE_VERSION=3.21 --build-arg GO_PROXY=https://proxy.golang.org .
```

### Chocolatey

```powershell
choco install ocm-cli
```

see: [chocolatey community package: ocm-cli](https://community.chocolatey.org/packages/ocm-cli)

## Examples

An example of how to use the `ocm` CLI in a Makefile can be found in [`examples/make`](examples/make/Makefile).

More comprehensive examples can be taken from the [`components`](components) contained in this repository. [Here](components/helmdemo/README.md) a complete component build including a multi-arch image is done and finally packaged into a CTF archive which can be transported into an OCI repository. See the readme files for details.

## GPG Public Key

The authenticity of released packages that have been uploaded to public repositories can be verified using our GPG public key. You can find the current key in the file [OCM-RELEASES-PUBLIC-CURRENT.gpg](https://ocm.software/gpg/OCM-RELEASES-PUBLIC-CURRENT.gpg) on our website. You can find the old keys in the website github repository [here](https://github.com/open-component-model/ocm-website/tree/main/static/gpg).

## Contributing

Code contributions, feature requests, bug reports, and help requests are very welcome. Please refer to the [Contributing Guide in the Community repository](https://github.com/open-component-model/community/blob/main/CONTRIBUTING.md) for more information on how to contribute to OCM.

OCM follows the [CNCF Code of Conduct](https://github.com/cncf/foundation/blob/main/code-of-conduct.md).

## Release Process

The release process is automated through a [github action workflow](https://github.com/open-component-model/ocm/actions/workflows/release.yaml).
Please refer to the [Release Process Documentation](RELEASE_PROCESS.md) for more information.

## Licensing

Copyright 2025 SAP SE or an SAP affiliate company and Open Component Model contributors.
Please see our [LICENSE](LICENSE) for copyright and license information.
Detailed information including third-party components and their licensing/copyright information is available [via the REUSE tool](https://api.reuse.software/info/github.com/open-component-model/ocm).
