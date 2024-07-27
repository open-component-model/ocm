# Open Component Model

[![OpenSSF Best Practices](https://bestpractices.coreinfrastructure.org/projects/7156/badge)](https://bestpractices.coreinfrastructure.org/projects/7156)
[![REUSE status](https://api.reuse.software/badge/github.com/open-component-model/ocm)](https://api.reuse.software/info/github.com/open-component-model/ocm)
[![OCM Integration Tests](https://github.com/open-component-model/ocm-integrationtest/actions/workflows/integrationtest.yaml/badge.svg?branch=main)](https://open-component-model.github.io/ocm-integrationtest/report.html)
[![Go Report Card](https://goreportcard.com/badge/github.com/open-component-model/ocm)](https://goreportcard.com/report/github.com/open-component-model/ocm)

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

## OCM CLI

The [`ocm` CLI](docs/reference/ocm.md) may also be used to interact with OCM mechanisms. It makes it easy to create component versions and embed them in build processes.

The `ocm` CLI documentation can be found [here](<(https://github.com/open-component-model/ocm/blob/main/docs/reference/ocm.md)>).

The code for the CLI can be found in [package `cmds/ocm`](https://github.com/open-component-model/ocm/blob/main/cmds/ocm).

The OCI and OCM support can be found in packages
[`pkg/contexts/oci`](pkg/contexts/oci) and [`pkg/contexts/ocm`](pkg/contexts/ocm).

## Installation

Install the latest release from any of

- [Homebrew](https://brew.sh)
- [Nix](https://nixos.org)
- [Docker](https://www.docker.com/)
- [Podman](https://podman.io/)
- [GitHub Releases](https://github.com/open-component-model/ocm/releases)

### Bash

To install with `bash` for macOS or Linux execute the following command:

```sh
curl -s https://ocm.software/install.sh | sudo bash
```

### Install using Homebrew

```sh
# Homebrew (macOS and Linux)
brew install open-component-model/tap/ocm
```

### Install using Nix (with [Flakes](https://nixos.wiki/wiki/Flakes))

```bash
# Nix (macOS, Linux, and Windows)
# ad hoc cmd execution
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

### Usage via Docker / Podman

```sh
podman run -t ghcr.io/open-component-model/ocm:latest --help
```

#### Build and run it yourself

```sh
podman build -t ocm .
podman run --rm -t ocm --loglevel debug --help
```

or interactively:

```sh
podman run --rm -it ocm /bin/sh
```

You can pass in the following arguments to override the predefined defaults:

- `GO_VERSION`: The **golang** version to be used for compiling.
- `ALPINE_VERSION`: The **alpine** version to be used as the base image.
- `GO_PROXY`: Your **go** proxy to be used for fetching dependencies.

Please check [hub.docker.com](https://hub.docker.com/_/golang/tags?page=1&name=alpine) for possible version combinations.

```sh
podman build -t ocm --build-arg GO_VERSION=1.22 --build-arg ALPINE_VERSION=3.19 --build-arg GO_PROXY=https://proxy.golang.org .
```

## Examples

An example of how to use the `ocm` CLI in a Makefile can be found in [`examples/make`](https://github.com/open-component-model/ocm/blob/main/examples/make/Makefile).

More comprehensive examples can be taken from the [`components`](https://github.com/open-component-model/ocm/tree/main/components) contained in this repository. [Here](components/helmdemo/README.md) a complete component build including a multi-arch image is done and finally packaged into a CTF archive which can be tranported into an OCI repository. See the readme files for details.

## Contributing

Code contributions, feature requests, bug reports, and help requests are very welcome. Please refer to the [Contributing Guide in the Community repository](https://github.com/open-component-model/community/blob/main/CONTRIBUTING.md) for more information on how to contribute to OCM.

OCM follows the [CNCF Code of Conduct](https://github.com/cncf/foundation/blob/main/code-of-conduct.md).

## Licensing

Copyright 2024 SAP SE or an SAP affiliate company and Open Component Model contributors.
Please see our [LICENSE](LICENSE) for copyright and license information.
Detailed information including third-party components and their licensing/copyright information is available [via the REUSE tool](https://api.reuse.software/info/github.com/open-component-model/ocm).
