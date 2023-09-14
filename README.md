# Open Component Model

[![OpenSSF Best Practices](https://bestpractices.coreinfrastructure.org/projects/7156/badge)](https://bestpractices.coreinfrastructure.org/projects/7156)
[![REUSE status](https://api.reuse.software/badge/github.com/open-component-model/ocm)](https://api.reuse.software/info/github.com/open-component-model/ocm)
[![OCM Integration Tests](https://github.com/open-component-model/ocm-integrationtest/actions/workflows/integrationtest.yaml/badge.svg?branch=main)](https://open-component-model.github.io/ocm-integrationtest/report.html)
[![Go Report Card](https://goreportcard.com/badge/github.com/open-component-model/ocm)](https://goreportcard.com/report/github.com/open-component-model/ocm)

The Open Component Model (OCM) is an open standard to describe software bills of delivery (SBOD). OCM is a technology-agnostic and machine-readable format focused on the software artifacts that must be delivered for software products.

OCM describes delivery [artifacts](docs/ocm/model.md#artifacts) that can be accessed from many types of [component repositories](docs/ocm/model.md#repositories).

Check out the [the main OCM project web page](https://ocm.software) to find out more. It is your central entry point to all kind of ocm related [docs and guides](https://ocm.software/docs/overview/context), the [spec](https://ocm.software/spec/) and all project related [github repositories](https://github.com/open-component-model). It also offers a [Getting Started](https://ocm.software/docs/guides/getting-started-with-ocm) to quickly make your hands dirty with ocm, its toolset and concepts :-)

## OCM Specifications

OCM defines a set of semantic, formatting, and other types of specifications that can be found in the [`ocm-spec` repository](https://github.com/open-component-model/ocm-spec). Start learning about the core concepts of OCM elements [here](https://github.com/open-component-model/ocm-spec/tree/main/doc/specification/elements).

## OCM Library

This project provides a Go library containing an API for interacting with the
[Open Component Model (OCM)](https://github.com/open-component-model/ocm-spec) elements and mechanisms.

The library currently supports the following [repository mappings](docs/ocm/interoperability.md):

- **OCI**: Use the repository prefix path of an OCI repository to implement an OCM
  repository.
- **CTF (Common Transport Format)**: Use a file-based binding to represent any set of
  component versions as filesystem content (directory, tar, tgz).
- **Component Archive**: Compose the content of a component version on the
  filesystem.

For the usage of the library to access OCM repositories, handle configuratio and credentials see the [examples section](examples/lib/README.md).

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

## Examples
An example of how to use the `ocm` CLI in a Makefile can be found in [`examples/make`](https://github.com/open-component-model/ocm/blob/main/examples/make/Makefile).

More comprehensive examples can be taken from the [`components`](https://github.com/open-component-model/ocm/tree/main/components) contained in this repository. [Here](components/helmdemo/README.md) a complete component build including a multi-arch image is done and finally packaged into a CTF archive which can be tranported into an OCI repository. See the readme files for details.


## Contributing

Code contributions, feature requests, bug reports, and help requests are very welcome. Please refer to the [Contributing Guide in the Community repository](https://github.com/open-component-model/community/blob/main/CONTRIBUTING.md) for more information on how to contribute to OCM.

OCM follows the [CNCF Code of Conduct](https://github.com/cncf/foundation/blob/main/code-of-conduct.md).

## Licensing

Copyright 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
Please see our [LICENSE](LICENSE) for copyright and license information.
Detailed information including third-party components and their licensing/copyright information is available [via the REUSE tool](https://api.reuse.software/info/github.com/open-component-model/ocm).
