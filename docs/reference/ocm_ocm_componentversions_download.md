## ocm ocm componentversions download &mdash; Download Ocm Component Versions

### Synopsis

```
ocm ocm componentversions download [<options>] {<components>} 
```

### Options

```
  -h, --help             help for download
  -O, --outfile string   output file or directory
      --repo string      repository name or spec
  -t, --type string      archive format (directory, tar, tgz) (default "directory")
```

### Description


Download component versions from an OCM repository. The result is stored in
component archives.

The files are named according to the component version name.

If the <code>--repo</code> option is specified, the given names are interpreted
relative to the specified repository using the syntax

<center>
    <pre>&lt;component>[:&lt;version>]</pre>
</center>

If no <code>--repo</code> option is specified the given names are interpreted 
as located OCM component version references:

<center>
    <pre>[&lt;repo type>::]&lt;host>[:&lt;port>][/&lt;base path>]//&lt;component>[:&lt;version>]</pre>
</center>

Additionally there is a variant to denote common transport archives
and general repository specifications

<center>
    <pre>[&lt;repo type>::]&lt;filepath>|&lt;spec json>[//&lt;component>[:&lt;version>]]</pre>
</center>

The <code>--repo</code> option takes an OCM repository specification:

<center>
    <pre>[&lt;repo type>::]&lt;configured name>|&lt;file path>|&lt;spec json></pre>
</center>

For the *Common Transport Format* the types <code>directory</code>,
<code>tar</code> or <code>tgz</code> is possible.

Using the JSON variant any repository type supported by the 
linked library can be used:

Dedicated OCM repository types:
- `ComponentArchive`

OCI Repository types (using standard component repository to OCI mapping):
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

* [ocm ocm componentversions](ocm_ocm_componentversions.md)	 &mdash; Commands acting on components
* [ocm ocm](ocm_ocm.md)	 &mdash; Dedicated command flavors for the Open Component Model
* [ocm](ocm.md)	 &mdash; Open Component Model command line client

