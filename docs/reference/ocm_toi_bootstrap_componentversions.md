## ocm toi bootstrap componentversions &mdash; Bootstrap Component Version

### Synopsis

```
ocm toi bootstrap componentversions [<options>] <action> {<component-reference>} {<resource id field>}
```

### Options

```
  -c, --credentials string   credentials file
  -h, --help                 help for componentversions
  -o, --outputs string       output file/directory
  -p, --parameters string    parameter file
```

### Description


Use the simple TOI bootstrap mechanism to execute actions for a TOI package resource
based on the content of an OCM component version and some command input describing
the dedicated installation target.

The package resource must have the type <code>toiPackage</code>.
This is a simple YAML file resource describing the bootstrapping of a dedicated kind
of software. See also the topic [ocm toi toi-bootstrapping](ocm_toi_toi-bootstrapping.md).

THis resource finally describes an executor image, which will be executed in a
container with the installation source and (instance specific) user settings.
The container is just executed, the framework make no assumption about the
meaning/outcome of the execution. Therefore, any kind of actions can be described and
issued this way, not on installation handling.

The first matching resource of this type is selected. Optionally a set of
identity attribute can be specified used to refine the match. This can be the
resource name and/or other key/value pairs (<code>&lt;attr>=&lt;value></code>).

If no output file is provided, the yaml representation of the outputs are
printed to standard out. If the output file is a directory, for every output a
dedicated file is created, otherwise the yaml representation is stored to the
file.

If no credentials file name is provided (option -c) the file 
<code>TOICredentials</code> is used, if present. If no parameter file name is
provided (option -p) the file <code>TOIParameters</code> is used, if present.

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

If a component lookup for building a reference closure is required
the <code>--lookup</code>  option can be used to specify a fallback
lookup repository. 
By default the component versions are searched in the repository
holding the component version for which the closure is determined.
For *Component Archives* this is never possible, because it only
contains a single component version. Therefore, in this scenario
this option must always be specified to be able to follow component
references.


### Examples

```

$ ocm toi bootstrap componentversion ghcr.io/mandelsoft/ocmdemoinstaller:0.0.1-dev

```

### SEE ALSO

##### Parents

* [ocm toi bootstrap](ocm_toi_bootstrap.md)	 &mdash; bootstrap components
* [ocm toi](ocm_toi.md)	 &mdash; Dedicated command flavors for the TOI layer
* [ocm](ocm.md)	 &mdash; Open Component Model command line client



##### Additional Help Topics

* [ocm toi bootstrap componentversions <b>toi-bootstrapping</b>](ocm_toi_bootstrap_componentversions_toi-bootstrapping.md)	 &mdash; Tiny OCM Installer based on component versions

