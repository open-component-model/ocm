# Open Component Model

This project provides a go library binding for working with the
Open Component Model (OCM) and an [OCM command line client](docs/reference/ocm.md).

The library supports an extensible set of repository bindings for OCM repositories:
- OCI: use a repository prefix path of an OCI repository to implement an OCM
  repository
- CTF (Common Transport Format): a file based binding to represent any set of
  component versions as file system content (directory, tar, tgz)
- Component Archive: Compose the content of a component version on the
  filesystem

Additionally it provides a generic solution
- to sign component version in any supported OCM repository implementation and
  verify signatures based on public keys or verified certificates.
- to transport component versions, per reference or as value among any of those 
  repository implementations.

This functionally is additionally put into a command line tool
([package `cmds/ocm`](cmds/ocm)), the 
[`ocm` tool](docs/reference/ocm.md), which provides the
most of the functionality of the library on the command line. This makes is easy
to embed the creation of component versions in build processes, for example in a 
[*makefile*](examples/make/Makefile).

The OCI and OCM support can be found in packages
[`pkg/contexts/oci`](pkg/contexts/oci) and [`pkg/contexts/ocm`](pkg/contexts/ocm).


There are several specifications:
 - [Naming Schemes](docs/names/README.md)
 - [Element Specifications](docs/formats/README.md)
