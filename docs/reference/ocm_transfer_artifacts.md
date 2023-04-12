## ocm transfer artifacts &mdash; Transfer OCI Artifacts

### Synopsis

```
ocm transfer artifacts [<options>] {<artifact-reference>} <target>
```

### Options

```
  -h, --help          help for artifacts
      --repo string   repository name or spec
  -R, --repo-name     transfer repository name
```

### Description


Transfer OCI artifacts from one registry to another one.
Several transfer scenarios are supported:
- copy a set of artifacts (for the same repository) into another registry
- copy a set of artifacts (for the same repository) into another repository
- copy artifacts from multiple repositories into another registry
- copy artifacts from multiple repositories into another registry with a given repository prefix (option -R)

By default the target is seen as a single repository if a repository is specified.
If a complete registry is specified as target, option -R is implied, but the source
must provide a repository. THis combination does not allow an artifact set as source, which
specifies no repository for the artifacts.

Sources may be specified as
- dedicated artifacts with repository and version or tag
- repository (without version), which is resolved to all available tags
- registry, if the specified registry implementation supports a namespace/repository lister,
  which is not the case for registries conforming to the OCI distribution specification.

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


### Examples

```
$ ocm oci artifact transfer ghcr.io/mandelsoft/kubelink:v1.0.0 gcr.io
$ ocm oci artifact transfer ghcr.io/mandelsoft/kubelink gcr.io
$ ocm oci artifact transfer ghcr.io/mandelsoft/kubelink gcr.io/my-project
$ ocm oci artifact transfer /tmp/ctf gcr.io/my-project
```

### SEE ALSO

##### Parents

* [ocm transfer](ocm_transfer.md)	 &mdash; Transfer artifacts or components
* [ocm](ocm.md)	 &mdash; Open Component Model command line client

