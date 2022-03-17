## ocm add resources

add resources to a component version

### Synopsis


Add resources specified in a resource file to a component version.
So far only component archives are supported as target.


```
ocm add resources [<options>] <target> {<resourcefile> | <var>=<value>} [flags]
```

### Options

```
      --addenv                 access environment for templating
  -h, --help                   help for resources
  -s, --settings stringArray   settings file with variable settings (yaml)
      --templater string       templater to use (subst, spiff, go) (default "subst")
```

### SEE ALSO

* [ocm add](ocm_add.md)	 - 

