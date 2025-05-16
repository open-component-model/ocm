---
title: "ocm get artifacts - Get Artifact Version"
linkTitle: "get artifacts"
url: "/docs/cli-reference/get/artifacts/"
sidebar:
  collapsed: true
menu:
  docs:
    name: "get artifacts"
---

### Synopsis

```bash
ocm get artifacts [<options>] {<artifact-reference>}
```

#### Aliases

```text
artifacts, artifact, art, a
```

### Options

```text
  -a, --attached           show attached artifacts
  -h, --help               help for artifacts
  -o, --output string      output mode (JSON, json, tree, wide, yaml)
  -r, --recursive          follow index nesting
      --repo string        repository name or spec
  -s, --sort stringArray   sort fields
```

### Description

Get lists all artifact versions specified, if only a repository is specified
all tagged artifacts are listed.
	

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



With the option <code>--recursive</code> the complete reference tree of a index is traversed.

With the option <code>--output</code> the output mode can be selected.
The following modes are supported:
  - <code></code> (default)
  - <code>JSON</code>
  - <code>json</code>
  - <code>tree</code>
  - <code>wide</code>
  - <code>yaml</code>

### Examples

```bash
$ ocm get artifact ghcr.io/open-component-model/ocm/component-descriptors/ocm.software/ocmcli
$ ocm get artifact ghcr.io/open-component-model/ocm/ocm.software/ocmcli/ocmcli-image:0.17.0
```

### SEE ALSO

#### Parents

* [ocm get](ocm_get.md)	 &mdash; Get information about artifacts and components
* [ocm](ocm.md)	 &mdash; Open Component Model command line client

