# Open Component Model

The Open Component Model provides a standard for describing delivery [artefacts](https://github.com/gardener/ocm-spec/tree/main/doc/specification/elements#artifacts) that can be accessed from many types of [component repositories](https://github.com/gardener/ocm-spec/tree/main/doc/specification/elements#repositories).

## OCM Specifications
OCM defines a set of semantic, formatting, and other types of specifications that can be found in the [`ocm-spec` repository](https://github.com/gardener/ocm-spec/tree/main/doc/specification). Start learning about the core concepts of OCM elements [here](https://github.com/gardener/ocm-spec/tree/main/doc/specification/elements).

## OCM Library
This project provides a Go library containing an API for interacting with the
[Open Component Model (OCM)](https://github.com/gardener/ocm-spec) elements and mechanisms.

The library currently supports the following [repository mappings](docs/ocm/interoperability.md):
- **OCI**: Use the repository prefix path of an OCI repository to implement an OCM
  repository.
- **CTF (Common Transport Format)**: Use a file-based binding to represent any set of
  component versions as filesystem content (directory, tar, tgz).
- **Component Archive**: Compose the content of a component version on the
  filesystem.

Additionally, OCM provides a generic solution for how to:
- Sign component versions in any supported OCM repository implementation.
- Verify signatures based on public keys or verified certificates.
- Transport component versions, per reference or as values to any of the
  repository implementations.

## OCM CLI
The [`ocm` CLI](docs/reference/ocm.md) may also be used to interact with OCM mechanisms. It makes it easy to create component versions and embed them in build processes.

The `ocm` CLI documentation can be found [here]((https://github.com/gardener/ocm/blob/main/docs/reference/ocm.md)).

The code for the CLI can be found in [package `cmds/ocm`](https://github.com/gardener/ocm/blob/main/cmds/ocm).

An example of how to use the `ocm` CLI in a Makefile can be found in [`examples/make`](https://github.com/gardener/ocm/blob/main/examples/make/Makefile).

The OCI and OCM support can be found in packages
[`pkg/contexts/oci`](pkg/contexts/oci) and [`pkg/contexts/ocm`](pkg/contexts/ocm).

## Contributing

Code contributions, feature requests, bug reports, and help requests are very welcome. Please refer to the [Contributing Guide in the Community repository](https://github.com/open-component-model/community/blob/main/CONTRIBUTING.md) for more information on how to contribute to OCM.

OCM follows the [CNCF Code of Conduct](https://github.com/cncf/foundation/blob/main/code-of-conduct.md).