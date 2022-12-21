## ocm describe artifacts &mdash; Describe Artifact Version

### Synopsis

```
ocm describe artifacts [<options>] {<artifact-reference>}
```

### Options

```
  -h, --help            help for artifacts
      --layerfiles      list layer files
  -o, --output string   output mode (JSON, json, yaml)
      --repo string     repository name or spec
```

### Description


Describe lists all artifact versions specified, if only a repository is specified
all tagged artifacts are listed.
Per version a detailed, potentially recursive description is printed.


If the repository/registry option is specified, the given names are interpreted
relative to the specified registry using the syntax

<center>
    <pre>&lt;OCI repository name>[:&lt;tag>][@&lt;digest>]</pre>
</center>

If no <code>--repo</code> option is specified the given names are interpreted 
as extended OCI artifact references.

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
- `ArtifactSet`
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
 - yaml


### Examples

```
$ ocm describe artifact ghcr.io/mandelsoft/kubelink
$ ocm describe artifact --repo OCIRegistry:ghcr.io mandelsoft/kubelink
```

### SEE ALSO

##### Parents

* [ocm describe](ocm_describe.md)	 &mdash; Describe various elements by using appropriate sub commands.
* [ocm](ocm.md)	 &mdash; Open Component Model command line client

