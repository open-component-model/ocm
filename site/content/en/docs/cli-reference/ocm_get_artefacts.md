
---
title: ocm_get_artefacts
url: /docs/cli-reference/ocm_get_artefacts/
date: 2022-08-12T11:14:49+01:00
draft: false
images: []
menu:
  docs:
    parent: cli-reference
toc: true
---
## ocm get artefacts &mdash; Get Artefact Version

### Synopsis

```
ocm get artefacts [<options>] {<artefact-reference>}
```

### Options

```
  -a, --attached           show attached artefacts
  -c, --closure            follow index nesting
  -h, --help               help for artefacts
  -o, --output string      output mode (JSON, json, tree, wide, yaml)
  -r, --repo string        repository name or spec
  -s, --sort stringArray   sort fields
```

### Description


Get lists all artefact versions specified, if only a repository is specified
all tagged artefacts are listed.
	
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

With the option <code>--closure</code> the complete reference tree of a index is traversed.

With the option <code>--output</code> the output mode can be selected.
The following modes are supported:
 - JSON
 - json
 - tree
 - wide
 - yaml


### Examples

```

$ ocm get artefact ghcr.io/mandelsoft/kubelink
$ ocm get artefact --repo OCIRegistry:ghcr.io mandelsoft/kubelink

```

### SEE ALSO

##### Parents

* [ocm get](ocm_get.md)	 - Get information about artefacts and components
* [ocm](ocm.md)	 - Open Component Model command line client

