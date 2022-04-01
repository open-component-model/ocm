## ocm get sources

get sources of a component version

### Synopsis


Get sources of a component version. Sources are specified
by identities. An identity consists of 
a name argument followed by optional <code>&lt;key>=&lt;value></code>
arguments.


```
ocm get sources [<options>]  <component> {<name> { <key>=<value> }} [flags]
```

### Options

```
  -c, --closure            follow component references
  -h, --help               help for sources
      --lookup string      repository name or spec for closure lookup fallback
  -o, --output string      output mode (wide, tree, yaml, json, JSON)
  -r, --repo string        repository name or spec
  -s, --sort stringArray   sort fields
```

### SEE ALSO

* [ocm get](ocm_get.md)	 - 

