## ocm download resources &mdash; Download Resources Of A Component Version

### Synopsis

```
ocm download resources [<options>]  <component> {<name> { <key>=<value> }}
```

### Options

```
  -c, --constraints constraints     version constraint
  -d, --download-handlers           use download handler if possible
      --downloader <name>=<value>   artifact downloader (<name>[:<artifact type>[:<media type>]]=<JSON target config) (default [])
  -x, --executable                  download executable for local platform
  -h, --help                        help for resources
      --latest                      restrict component versions to latest
      --lookup stringArray          repository name or spec for closure lookup fallback
  -O, --outfile string              output file or directory
  -r, --recursive                   follow component reference nesting
      --repo string                 repository name or spec
  -t, --type stringArray            resource type filter
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
    <pre>&lt;component>/&lt;version>{/&lt;nested component>/&lt;version>}</pre>
</center>

The resource files are named according to the resource identity in the
component descriptor. If this identity is just the resource name, this name
is ised. If additional identity attributes are required, this name is
append by a comma separated list of <code>&lt;name>=&lt;>value></code> pairs
separated by a "-" from the plain name. This attribute list is alphabetical
order:

<center>
    <pre>&lt;resource name>[-[&lt;name>=&lt;>value>]{,&lt;name>=&lt;>value>}]</pre>
</center>



If the option <code>--constraints</code> is given, and no version is specified for a component, only versions matching
the given version constraints (semver https://github.com/Masterminds/semver) are selected. With <code>--latest</code> only
the latest matching versions will be selected.


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

Using the JSON variant any repository types supported by the 
linked library can be used:

Dedicated OCM repository types:
  - <code>ComponentArchive</code>: v1

OCI Repository types (using standard component repository to OCI mapping):
  - <code>ArtifactSet</code>: v1
  - <code>CommonTransportFormat</code>: v1
  - <code>DockerDaemon</code>: v1
  - <code>Empty</code>: v1
  - <code>OCIRegistry</code>: v1
  - <code>oci</code>: v1
  - <code>ociRegistry</code>



If the <code>--downloader</code> option is specified, appropriate downloader handlers
are configured for the operation. It has the following format

<center>
    <pre>&lt;name>:&lt;artifact type>:&lt;media type>=&lt;yaml target config></pre>
</center>

The downloader name may be a path expression with the following possibilities:
  - <code>plugin</code>: [downloaders provided by plugins]
    
    sub namespace of the form <code>&lt;plugin name>/&lt;handler></code>

  - <code>ocm/dirtree</code>: downloading directory tree-like resources
    
    The <code>dirtree</code> downloader is able to to download directory-tree like
    resources as directory stricture (default) or archive.
    The following artifact media types are supported:
      - <code>application/vnd.oci.image.manifest.v1+tar+gzip</code>
      - <code>application/x-tgz</code>
      - <code>application/x-tar+gzip</code>
      - <code>application/x-tar</code>
    
    By default, it is registered for the following resource types:
      - <code>directoryTree</code>
      - <code>filesystem</code>
    
    If accepts a config with the following fields:
      - <code>asArchive</code>: flag to request an archive download
      - <code>configTypes</code>: a list of accepted OCI config archive types
        defaulted by <code>application/vnd.oci.image.config.v1+json</code>.



See [ocm ocm-downloadhandlers](ocm_ocm-downloadhandlers.md) for further details on using
download handlers.



The library supports some downloads with semantics based on resource types. For example a helm chart
can be download directly as helm chart archive, even if stored as OCI artifact.
This is handled by download handler. Their usage can be enabled with the <code>--download-handlers</code>
option. Otherwise the resource as returned by the access method is stored.


With the option <code>--recursive</code> the complete reference tree of a component reference is traversed.

\
If a component lookup for building a reference closure is required
the <code>--lookup</code>  option can be used to specify a fallback
lookup repository. By default, the component versions are searched in
the repository holding the component version for which the closure is
determined. For *Component Archives* this is never possible, because
it only contains a single component version. Therefore, in this scenario
this option must always be specified to be able to follow component
references.


### SEE ALSO

##### Parents

* [ocm download](ocm_download.md)	 &mdash; Download oci artifacts, resources or complete components
* [ocm](ocm.md)	 &mdash; Open Component Model command line client



##### Additional Links

* [<b>ocm ocm-downloadhandlers</b>](ocm_ocm-downloadhandlers.md)	 &mdash; List of all available download handlers

