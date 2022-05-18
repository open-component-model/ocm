## ocm download artefacts

download oci artefacts

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

<center><code>&lt;OCI repository name>[:&lt;tag>][@&lt;digest>]</code></center>

If no <code>--repo</code> option is specified the given names are interpreted 
as extended CI artefact references.

<center><code>[&lt;repo type>::]&lt;host>[:&lt;port>]/&lt;OCI repository name>[:&lt;tag>][@&lt;digest>]</code></center>

The <code>--repo</code> option takes a repository/OCI registry specification:

<center><code>[&lt;repo type>::]&lt;configured name>|&lt;file path>|&lt;spec json></code></center>

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

