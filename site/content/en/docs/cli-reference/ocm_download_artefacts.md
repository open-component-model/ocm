
---
title: ocm_download_artefacts
url: /docs/cli-reference/ocm_download_artefacts/
date: 2022-08-12T11:14:49+01:00
draft: false
images: []
menu:
  docs:
    parent: cli-reference
toc: true
---
## ocm download artefacts &mdash; Download Oci Artefacts

### Synopsis

```
ocm download artefacts [<options>]  {<artefact>} 
```

### Options

```
  -h, --help             help for artefacts
  -O, --outfile string   output file or directory
  -r, --repo string      repository name or spec
  -t, --type string      archive format (default "directory")
```

### Description


Download artefacts from an OCI registry. The result is stored in
artefact set format, without the repository part

The files are named according to the artefact repository name.

If the repository/registry option is specified, the given names are interpreted
relative to the specified registry using the syntax

<center>
    <pre>&lt;OCI repository name>[:&lt;tag>][@&lt;digest>]</pre>
</center>

If no <code>--repo</code> option is specified the given names are interpreted 
as extended CI artefact references.

<center>
    <pre>[&lt;repo type>::]&lt;host>[:&lt;port>]/&lt;OCI repository name>[:&lt;tag>][@&lt;digest>]</pre>
</center>

The <code>--repo</code> option takes a repository/OCI registry specification:

<center>
    <pre>[&lt;repo type>::]&lt;configured name>|&lt;file path>|&lt;spec json></pre>
</center>

For the *Common Transport Format* the types <code>directory</code>,
<code>tar</code> or <code>tgz</code> are possible.

Using the JSON variant any repository type supported by the 
linked library can be used:
- `ArtefactSet`
- `CommonTransportFormat`
- `DockerDaemon`
- `Empty`
- `OCIRegistry`
- `oci`
- `ociRegistry`

The <code>--type</code> option accepts a file format for the
target archive to use. The following formats are supported:
- directory
- tar
- tgz
The default format is <code>directory</code>.


### SEE ALSO

##### Parents

* [ocm download](ocm_download.md)	 - Download oci artefacts, resources or complete components
* [ocm](ocm.md)	 - Open Component Model command line client

