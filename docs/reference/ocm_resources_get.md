## ocm resources get

get resources of a component version

### Synopsis


Get resources of a component version. Sources are specified
by identities. An identity consists of 
a name argument followed by optional <code>&lt;key>=&lt;value></code>
arguments.


```
ocm resources get [<options>]  <component> {<name> { <key>=<value> }} [flags]
```

### Options

```
  -c, --closure            follow component references
  -h, --help               help for get
      --lookup string      repository name or spec for closure lookup fallback
  -o, --output string      output mode (wide, tree, yaml, json, JSON)
  -r, --repo string        repository name or spec
  -s, --sort stringArray   sort fields
```

### SEE ALSO

* [ocm resources](ocm_resources.md)	 - 

