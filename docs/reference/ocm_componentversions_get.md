## ocm componentversions get &mdash; Get Component Version

### Synopsis

```
ocm componentversions get [<options>] {<component-reference>}
```

### Options

```
  -c, --closure            follow component reference nesting
  -h, --help               help for get
  -o, --output string      output mode (JSON, json, tree, wide, yaml)
  -r, --repo string        repository name or spec
  -S, --scheme string      schema version
  -s, --sort stringArray   sort fields
```

### Description


Get lists all component versions specified, if only a component is specified
all versions are listed.

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

With the option <code>--closure</code> the complete reference tree of a component reference is traversed.

It the option <code>--scheme</code> is given, the given component descriptor is converted to given format for output.
The following schema versions are supported:

  - <code>ocm.gardener.cloud/v3alpha1</code>: 

  - <code>v2</code>: 


With the option <code>--output</code> the output mode can be selected.
The following modes are supported:
 - JSON
 - json
 - tree
 - wide
 - yaml


### Examples

```

$ ocm get componentversion ghcr.io/mandelsoft/kubelink
$ ocm get componentversion --repo OCIRegistry:ghcr.io mandelsoft/kubelink

```

### SEE ALSO

##### Parents

* [ocm componentversions](ocm_componentversions.md)	 &mdash; Commands acting on components
* [ocm](ocm.md)	 &mdash; Open Component Model command line client

