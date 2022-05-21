## ocm resources get &mdash; Get Resources Of A Component Version

### Synopsis

```
ocm resources get [<options>]  <component> {<name> { <key>=<value> }}
```

### Options

```
  -c, --closure            follow component reference nesting
  -h, --help               help for get
      --lookup string      repository name or spec for closure lookup fallback
  -o, --output string      output mode (JSON, json, tree, wide, yaml)
  -r, --repo string        repository name or spec
  -s, --sort stringArray   sort fields
```

### Description


Get resources of a component version. Reources are specified
by identities. An identity consists of 
a name argument followed by optional <code>&lt;key>=&lt;value></code>
arguments.

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

With the option <code>--output</code> the out put mode can be selected.
The following modes are supported:
 - JSON
 - json
 - tree
 - wide
 - yaml


### SEE ALSO

##### Parents

* [ocm resources](ocm_resources.md)	 - Commands acting on component resources
* [ocm](ocm.md)	 - Open Component Model command line client

