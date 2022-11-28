## ocm add references &mdash; Add Aggregation Information To A Component Version

### Synopsis

```
ocm add references [<options>] [<target>] {<referencefile> | <var>=<value>}
```

### Options

```
      --addenv                 access environment for templating
  -F, --file string            target file/directory (default "component-archive")
  -h, --help                   help for references
  -s, --settings stringArray   settings file with variable settings (yaml)
      --templater string       templater to use (subst, spiff, go) (default "subst")
```


#### Reference Meta Data Options

```
      --component string       component name
      --extra <name>=<value>   reference extra identity
      --label <name>=<YAML>    reference label (leading * indicates signature relevant, optional version separated by @)
      --name string            reference name
      --reference YAML         reference meta data (yaml)
      --version string         reference version
```

### Description


Add aggregation information specified in a reference file to a component version.
So far only component archives are supported as target.

This command accepts reference specification files describing the references
to add to a component version. Elements must follow the reference meta data
description scheme of the component descriptor.

It is possible to describe a single reference via command line options.
The meta data of this element is described by the argument of option <code>--reference</code>,
which must be a YAML or JSON string.
Alternatively, the <em>name</em> and <em>version</em> can be specified with the
options <code>--name</code> and <code>--version</code>. With the option <code>--extra</code>
it is possible to add extra identity attributes. Explicitly specified options
override values specified by the <code>--reference</code> option.
(Note: Go templates are not supported for YAML-based option values. Besides
this restriction, the finally composed element description is still processd
by the selected templater.) 

The component name can be specified with the option <code>--component</code>. 
Therefore, basic references not requiring any additional labels or extra
identities can just be specified by those simple value options without the need
for the YAML option.

All yaml/json defined resources can be templated.
Variables are specified as regular arguments following the syntax <code>&lt;name>=&lt;value></code>.
Additionally settings can be specified by a yaml file using the <code>--settings <file></code>
option. With the option <code>--addenv</code> environment variables are added to the binding.
Values are overwritten in the order environment, settings file, command line settings. 

Note: Variable names are case-sensitive.

Example:
<pre>
&lt;command> &lt;options> -- MY_VAL=test &lt;args>
</pre>

There are several templaters that can be selected by the <code>--templater</code> option:
- <code>go</code> go templating supports complex values.

  <pre>
    key:
      subkey: "abc {{.MY_VAL}}"
  </pre>
  
- <code>none</code> do not do any substitution.

- <code>spiff</code> [spiff templating](https://github.com/mandelsoft/spiff).

  It supports complex values. the settings are accessible using the binding <code>values</code>.
  <pre>
    key:
      subkey: "abc (( values.MY_VAL ))"
  </pre>
  
- <code>subst</code> simple value substitution with the <code>drone/envsubst</code> templater.

  It supports string values, only. Complex settings will be json encoded.
  <pre>
    key:
      subkey: "abc ${MY_VAL}"
  </pre>
  


### Examples



Add a reference directly by options

<pre>
$ ocm add references &dash;&dash;file path/to/ca &dash;&dash;name myref &dash;&dash;component github.com/my/component &dash;&dash;version ${VERSION}
</pre>


Add a reference by a description file:

*references.yaml*:

<pre>
&dash;&dash;&dash;
name: myref
component: github.com/my/component
version: ${VERSION]
</pre>


<pre>
$ ocm add references  path/to/ca  references.yaml VERSION=1.0.0
</pre>



### SEE ALSO

##### Parents

* [ocm add](ocm_add.md)	 &mdash; Add resources or sources to a component archive
* [ocm](ocm.md)	 &mdash; Open Component Model command line client

