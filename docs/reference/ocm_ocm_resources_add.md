## ocm ocm resources add

add source information to a component version

### Synopsis

```
ocm ocm resources add [<options>] <target> {<resourcefile> | <var>=<value>}
```

### Options

```
      --addenv                 access environment for templating
  -h, --help                   help for add
  -s, --settings stringArray   settings file with variable settings (yaml)
      --templater string       templater to use (subst, spiff, go) (default "subst")
```

### Description


Add resource information specified in a resource file to a component version.
So far only component archives are supported as target.

Templating:
All yaml/json defined resources can be templated.
Variables are specified as regular arguments following the syntax <code>&lt;name>=&lt;value></code>.
Additionally settings can be specified by a yaml file using the <code>--settings <file></code>
option. With the option <code>--addenv</code> environment variables are added to the binding.
Values are overwritten in the order environment, settings file, commmand line settings. 

Note: Variable names are case-sensitive.

Example:
<pre>
<command> <options> -- MY_VAL=test <args>
</pre>

There are several templaters that can be selected by the <code>--templater</code> option:
- envsubst: simple value substitution with the <code>drone/envsubst</code> templater. It
  supports string values, only. Complext settings will be json encoded.
  <pre>
  key:
    subkey: "abc ${MY_VAL}"
  </pre>

- go: go templating supports complex values.
  <pre>
  key:
    subkey: "abc {{.MY_VAL}}"
  </pre>

- spiff: [spiff templating](https://github.com/mandelsoft/spiff) supports
  complex values. the settings are accessible using the binding <tt>values</tt>.
  <pre>
  key:
    subkey: "abc (( values.MY_VAL ))"
  </pre>


### SEE ALSO

##### Parents

* [ocm ocm resources](ocm_ocm_resources.md)	 - Commands acting on component resources
* [ocm ocm](ocm_ocm.md)	 - Dedicated command flavors for the Open Component Model
* [ocm](ocm.md)	 - ocm command line client

