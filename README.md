# Open Component Model

This projects provides a go library binding for working with the
Open Component Model (OCM)

It supports an extensible set of repository bindings for OCM repositories:
- OCI: use a repository prefix path of an OCI repository to implement an OCM
  repository
- CTF (Common Transport Format): a file based binding to represent any set of
  component versions as file system content (directory, tar, tgz)
- Component Archive: Compose the content of a component version on the
  filesystem

Additionally it provides a generic solution
- sign component version in any support OCM repository implementation and verify
  signatures
- to transport component versions, per reference or as value among any of those 
  repository implementations.

This functionally is additionally put into a command line interface, the 
[`ocm` tool](docs/reference/ocm.md), which supports makes it easy to use the
complete functionality on the command line. This makes is easy to embed the
creation of component versions in build processes, for example in a 
[*makefile*](examples/make/Makefile).
