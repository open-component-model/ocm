## ocm download artifacts &mdash; Download Oci Artifacts

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
      --oci-layout       download as OCI Image Layout (blobs in blobs/<algorithm>/<encoded>)
  -O, --outfile string   output file or directory
      --repo string      repository name or spec
  -t, --type string      archive format (directory, tar, tgz) (default "directory")
```

### Description

Download artifacts from an OCI registry. The result is stored in
artifact set format, without the repository part.

The files are named according to the artifact repository name.

By default, blobs are stored in OCM artifact set format (blobs/<algorithm>.<encoded>).
Use --oci-layout to store blobs in OCI Image Layout format (blobs/<algorithm>/<encoded>)
for compatibility with tools that expect the OCI Image Layout Specification.


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

Option <code>--oci-layout</code> changes the blob storage structure in the downloaded
artifact. Without this option, blobs are stored in a flat directory at
<code>blobs/&lt;algorithm&gt;.&lt;encoded&gt;</code> (e.g., <code>blobs/sha256.abc123...</code>).
With this option, blobs are stored in a nested directory structure at
<code>blobs/&lt;algorithm&gt;/&lt;encoded&gt;</code> (e.g., <code>blobs/sha256/abc123...</code>)
as specified by the OCI Image Layout Specification
(see <a href="https://github.com/opencontainers/image-spec/blob/main/image-layout.md">
https://github.com/opencontainers/image-spec/blob/main/image-layout.md</a>).


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

