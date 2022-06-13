## ocm oci tags show &mdash; Show Dedicated Tags Of OCI Artefacts

### Synopsis

```
ocm oci tags show [<options>] <component> {<version pattern>}
```

### Options

```
  -h, --help          help for show
  -l, --latest        show only latest tags
  -r, --repo string   repository name or spec
  -o, --semantic      show semantic tags
  -s, --semver        show only semver compliant tags
```

### Description


Match tags of an artefact against some patterns.

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


### Examples

```

$ oci show tags ghcr.io/mandelsoft/kubelink

```

### SEE ALSO

##### Parents

* [ocm oci tags](ocm_oci_tags.md)	 - Commands acting on OCI tag names
* [ocm oci](ocm_oci.md)	 - Dedicated command flavors for the OCI layer
* [ocm](ocm.md)	 - Open Component Model command line client

