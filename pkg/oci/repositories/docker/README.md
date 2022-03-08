# Docker Daemon as OCI Repository

This package provides a mapping of the image repository behind a docker
daemon to the OCI registry access API.

This is only possible with a set of limitation:
- It is only possible to store and access flat images
- There is no access by digests, only by tags.
- The docker image id can be used as pseudo digest (without algorithm)
