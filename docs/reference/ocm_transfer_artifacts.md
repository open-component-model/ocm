---
title: "transfer artifacts"
menu:
  docs:
    parent: transfer
---
## ocm transfer artifacts &mdash; Transfer OCI Artifacts

### Synopsis

```bash
ocm transfer artifacts [<options>] {<artifact-reference>} <target>
```

#### Aliases

```text
artifacts, artifact, art, a
```

### Options

```text
  -h, --help          help for artifacts
      --repo string   repository name or spec
  -R, --repo-name     transfer repository name
```

### Description

Transfer OCI artifacts from one registry to another one.
Several transfer scenarios are supported:
- copy a set of artifacts (for the same repository) into another registry
- copy a set of artifacts (for the same repository) into another repository
- copy artifacts from multiple repositories into another registry
- copy artifacts from multiple repositories into another registry with a given repository prefix (option -R)

By default, the target is seen as a single repository if a repository is specified.
If a complete registry is specified as target, option -R is implied, but the source
must provide a repository. THis combination does not allow an artifact set as source, which
specifies no repository for the artifacts.

Sources may be specified as
- dedicated artifacts with repository and version or tag
- repository (without version), which is resolved to all available tags
- registry, if the specified registry implementation supports a namespace/repository lister,
  which is not the case for registries conforming to the OCI distribution specification.

Note that there is an indirection of "ocm oci artifact" to "ocm transfer artifact" out of convenience.

If the repository/registry option is specified, the given names are interpreted
relative to the specified registry using the syntax

<center>
    <pre>&lt;OCI repository name>[:&lt;tag>][@&lt;digest>]</pre>
</center>

If no <code>--repo</code> option is specified the given names are interpreted
as extended OCI artifact references.

<center>
    <pre>[&lt;repo type>::]&lt;host>[:&lt;port>]/&lt;OCI repository name>[:&lt;tag>][@&lt;digest>]</pre>
</center>

The <code>--repo</code> option takes a repository/OCI registry specification:

<center>
    <pre>[&lt;repo type>::]&lt;configured name>|&lt;file path>|&lt;spec json></pre>
</center>

For the *Common Transport Format* the types <code>directory</code>,
<code>tar</code> or <code>tgz</code> are possible.

Using the JSON variant any repository types supported by the
linked library can be used:
  - <code>ArtifactSet</code>: v1
  - <code>CommonTransportFormat</code>: v1
  - <code>DockerDaemon</code>: v1
  - <code>Empty</code>: v1
  - <code>OCIRegistry</code>: v1
  - <code>oci</code>: v1
  - <code>ociRegistry</code>

### Examples

```bash
# Simple:
$ ocm transfer artifact ghcr.io/open-component-model/ocm/ocm.software/ocmcli/ocmcli-image:0.17.0 ghcr.io/MY_USER/ocmcli:0.17.0
$ ocm transfer artifact ghcr.io/open-component-model/ocm/ocm.software/ocmcli/ocmcli-image ghcr.io/MY_USER/ocmcli
$ ocm transfer artifact ghcr.io/open-component-model/ocm/ocm.software/ocmcli/ocmcli-image gcr.io
$ ocm transfer artifact transfer /tmp/ctf ghcr.io/MY_USER/ocmcli

# Equivalent to ocm transfer artifact:
$ ocm oci artifact transfer

# Complex:
# Transfer an artifact from a CTF into an OCI Repository:
# 1. Get the link to all artifacts in the CTF with "ocm get artifact $PATH_TO_CTF",
$ ocm get artifact $PATH_TO_CTF
REGISTRY                                                               REPOSITORY
CommonTransportFormat::$PATH_TO_CTF/ component-descriptors/ocm.software/ocmcli
# 2. Then use any combination to form an artifact reference:
$ ocm transfer artifact  CommonTransportFormat::$PATH_TO_CTF//component-descriptors/ocm.software/ocmcli ghcr.io/open-component-model/ocm:latest
```

### SEE ALSO

#### Parents

* [ocm transfer](ocm_transfer.md)	 &mdash; Transfer artifacts or components
* [ocm](ocm.md)	 &mdash; Open Component Model command line client

