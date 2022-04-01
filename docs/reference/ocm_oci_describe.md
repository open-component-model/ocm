## ocm oci describe

describe artefact version

### Synopsis


Describe lists all artefact versions specified, if only a repository is specified
all tagged artefacts are listed.
Per version a detailed, potentially recursive description is printed.


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
- `CommonTransportFormat`
- `DockerDaemon`
- `Empty`
- `OCIRegistry`


*Example:*
<pre>
$ ocm describe artefact ghcr.io/mandelsoft/kubelink
$ ocm describe artefact --repo OCIRegistry:ghcr.io mandelsoft/kubelink
</pre>


```
ocm oci describe [<options>] {<artefact-reference>} [flags]
```

### Options

```
  -h, --help               help for describe
      --layerfiles         list layer files
  -o, --output string      output mode (yaml, json, JSON)
  -r, --repo string        repository name or spec
  -s, --sort stringArray   sort fields
```

### SEE ALSO

* [ocm oci](ocm_oci.md)	 - 

