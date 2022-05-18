## ocm download resources

download resources of a component version

### Synopsis

```
ocm download resources [<options>]  <component> {<name> { <key>=<value> }}
```

### Options

```
  -c, --closure             follow component reference nesting
  -d, --download-handlers   use download handler if possible
  -h, --help                help for resources
      --lookup string       repository name or spec for closure lookup fallback
  -O, --outfile string      output file or directory
  -r, --repo string         repository name or spec
```

### Description


Download resources of a component version. Resources are specified
by identities. An identity consists of 
a name argument followed by optional <code>&lt;key>=&lt;value></code>
arguments.

The option <code>-O</code> is used to declare the output destination.
For a single resource to download, this is the file written for the
resource blob. If multiple resources are selected, a directory structure
is written into the given directory for every involved component version
as follows:

<center>
<code>&lt;component>/&lt;version>{/&lt;nested component>/&lt;version>}</code>
</center>

The resource files are named according to the resource identity in the
component descriptor. If this identity is just the resource name, this name
is ised. If additional identity attributes are required, this name is
append by a comma separated list of <code>&lt;name>=&lt>value></code> pairs
separated by a "-" from the plain name. This attribute list is alphabetical
order:

<center>
<code>&lt;resource name>[-[&lt;name>=&lt>value>]{,&lt;name>=&lt>value>}]</code>
</center>


If the <code>--repo</code> option is specified, the given names are interpreted
relative to the specified repository using the syntax

<center><code>&lt;component>[:&lt;version>]</code></center>

If no <code>--repo</code> option is specified the given names are interpreted 
as located OCM component version references:

<center><code>[&lt;repo type>::]&lt;host>[:&lt;port>][/&lt;base path>]//&lt;component>[:&lt;version>]</code></center>

Additionally there is a variant to denote common transport archives
and general repository specifications

<center><code>[&lt;repo type>::]&lt;filepath>|&lt;spec json>[//&lt;component>[:&lt;version>]]</code></center>

The <code>--repo</code> option takes an OCM repository specification:

<center><code>[&lt;repo type>::]&lt;configured name>|&lt;file path>|&lt;spec json></code></center>

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

The library supports some downloads with semantics based on resource types. For example a helm chart
can be download directly as helm chart archive, even if stored as OCI artefact.
This is handled by download handler. Their usage can be enabled with the <code>--download-handlers</code>
option. Otherwise the resource as returned by the access method is stored.

With the option <code>--closure</code> the complete reference tree of a component reference is traversed.

If a component lookup for building a reference closure is required
the <code>--lookup</code>  option can be used to specify a fallback
lookup repository. 
By default the component versions are searched in the repository
holding the component version for which the closure is determined.
For *Component Archives* this is never possible, because it only
contains a single component version. Therefore, in this scenario
this option must always be specified to be able to follow component
references.


### SEE ALSO

##### Parents

* [ocm download](ocm_download.md)	 - Download oci artefacts, resources or complete components
* [ocm](ocm.md)	 - Open Component Model command line client

