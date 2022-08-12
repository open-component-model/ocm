
---
title: ocm_oci_artefacts_transfer
url: /docs/cli-reference/ocm_oci_artefacts_transfer/
date: 2022-08-12T11:14:49+01:00
draft: false
images: []
menu:
  docs:
    parent: cli-reference
toc: true
---
## ocm oci artefacts transfer &mdash; Transfer OCI Artefacts

### Synopsis

```
ocm oci artefacts transfer [<options>] {<artefact-reference>}
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

$ ocm oci transfer ghcr.io/mandelsoft/kubelink gcr.io

```

### SEE ALSO

##### Parents

* [ocm oci artefacts](ocm_oci_artefacts.md)	 &mdash; Commands acting on OCI artefacts
* [ocm oci](ocm_oci.md)	 &mdash; Dedicated command flavors for the OCI layer
* [ocm](ocm.md)	 &mdash; Open Component Model command line client

