## ocm get artefact

get artefact version

### Synopsis


Get lists all artefact versions specified, if only a repository is specified
all tagged artefacts are listed.

If no <code>repo</code> option is specified the given names are interpreted 
as located OCI artefact names. 

The options follows the syntax [<repotype>::]<repospec>. The following
repository types are supported yet:
- <code>OCIRegistry</code>: The given repository spec is used as base url

Without a specified type prefix any JSON representation of an OCI repository
specification supported by the OCM library or the name of an OCI repository
configured in the used config file can be used.

If the repository option is specified, the given artefact names are interpreted
relative to the specified repository.

*Example:*
<pre>
$ ocm get artefact ghcr.io/mandelsoft/kubelink
$ ocm get artefact --repo OCIRegistry:ghcr.io mandelsoft/kubelink
</pre>


```
ocm get artefact [<options>] {<artefact-reference>} [flags]
```

### Options

```
  -h, --help               help for artefact
  -o, --output string      output mode (wide, yaml, json, JSON)
  -r, --repo string        repository name or spec
  -s, --sort stringArray   sort fields
```

### SEE ALSO

* [ocm get](ocm_get.md)	 - 

