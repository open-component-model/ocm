## ocm get plugins &mdash; Get Plugins

### Synopsis

```
ocm get plugins [<options>] {<plugin name>}
```

### Options

```
  -h, --help               help for plugins
  -o, --output string      output mode (JSON, json, wide, yaml)
      --repo string        repository name or spec
  -s, --sort stringArray   sort fields
```

### Description


Get lists information for all plugins specified, if no plugin is specified
all registered ones are listed.

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

With the option <code>--output</code> the output mode can be selected.
The following modes are supported:
 - JSON
 - json
 - wide
 - yaml


### Examples

```
$ ocm get plugins
$ ocm get plugins demo -o yaml
```

### SEE ALSO

##### Parents

* [ocm get](ocm_get.md)	 &mdash; Get information about artefacts and components
* [ocm](ocm.md)	 &mdash; Open Component Model command line client

