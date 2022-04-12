## ocm oci transfer

transfer OCI artefacts

### Synopsis

```
ocm oci transfer [<options>] {<artefact-reference>}
```

### Options

```
  -h, --help          help for transfer
  -r, --repo string   repository name or spec
```

### Description


Transfer OCI artefacts from one registry to another one

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

$ ocm oci transfer ghcr.io/mandelsoft/kubelink gcr.io

```

### SEE ALSO

##### Parent

* [ocm oci](ocm_oci.md)	 - 

