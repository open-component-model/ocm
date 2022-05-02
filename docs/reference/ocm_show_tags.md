## ocm show tags

show dedicated tags of OCI artefacts

### Synopsis

```
ocm show tags [<options>] <component> {<version pattern>}
```

### Options

```
  -h, --help          help for tags
  -l, --latest        show only latest tags
  -r, --repo string   repository name or spec
  -o, --semantic      show semantic tags
  -s, --semver        show only semver compliant tags
```

### Description


Match tags of an artefact against some patterns.

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


### Examples

```

$ oci show tags ghcr.io/mandelsoft/kubelink

```

### SEE ALSO

##### Parents

* [ocm show](ocm_show.md)	 - Show tags or versions
* [ocm](ocm.md)	 - ocm command line client

