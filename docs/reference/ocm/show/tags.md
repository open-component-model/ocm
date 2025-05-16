---
title: "ocm show tags &mdash; Show Dedicated Tags Of OCI Artifacts"
linkTitle: "show tags"
url: "/docs/cli-reference/show/tags/"
sidebar:
  collapsed: true
---

### Synopsis

```bash
ocm show tags [<options>] <component> {<version pattern>}
```

### Options

```text
  -h, --help          help for tags
  -l, --latest        show only latest tags
      --repo string   repository name or spec
  -o, --semantic      show semantic tags
  -s, --semver        show only semver compliant tags
```

### Description

Match tags of an artifact against some patterns.


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
$ ocm show tags ghcr.io/open-component-model/ocm/ocm.software/ocmcli/ocmcli-image
$ ocm oci show tags ghcr.io/open-component-model/ocm/ocm.software/ocmcli/ocmcli-image
```

### SEE ALSO

#### Parents

* [ocm show](ocm_show.md)	 &mdash; Show tags or versions
* [ocm](ocm.md)	 &mdash; Open Component Model command line client

