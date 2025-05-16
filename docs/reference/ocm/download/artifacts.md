---
title: "ocm download artifacts - Download Oci Artifacts"
linkTitle: "download artifacts"
url: "/docs/cli-reference/download/artifacts/"
sidebar:
  collapsed: true
menu:
  docs:
    name: "download artifacts"
---

### Synopsis

```bash
ocm download artifacts [<options>]  {<artifact>}
```

#### Aliases

```text
artifacts, artifact, art, a
```

### Options

```text
      --dirtree          extract as effective filesystem content
  -h, --help             help for artifacts
      --layers ints      extract dedicated layers
  -O, --outfile string   output file or directory
      --repo string      repository name or spec
  -t, --type string      archive format (directory, tar, tgz) (default "directory")
```

### Description

Download artifacts from an OCI registry. The result is stored in
artifact set format, without the repository part

The files are named according to the artifact repository name.


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



With option <code>--layers</code> it is possible to request the download of
dedicated layers, only. Option <code>--dirtree</code> expects the artifact to
be a layered filesystem (for example OCI Image) and provided the effective
filesystem content.


The <code>--type</code> option accepts a file format for the
target archive to use. It is only evaluated if the target
archive does not exist yet. The following formats are supported:
- directory
- tar
- tgz

The default format is <code>directory</code>.

### SEE ALSO

#### Parents

* [ocm download](ocm_download.md)	 &mdash; Download oci artifacts, resources or complete components
* [ocm](ocm.md)	 &mdash; Open Component Model command line client

