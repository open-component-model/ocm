## ocm add componentversions &mdash; Add Component Version(S) To A (New) Transport Archive

### Synopsis

```
ocm add componentversions [<options>] [--version <version>] [<ctf archive>] {<components.yaml>}
```

##### Aliases

```
componentversions, componentversion, cv, components, component, comps, comp, c
```

### Options

```
      --addenv                 access environment for templating
  -C, --complete               include all referenced component version
  -L, --copy-local-resources   transfer referenced local resources by-value
  -V, --copy-resources         transfer referenced resources by-value
  -c, --create                 (re)create archive
      --dry-run                evaluate and print component specifications
  -F, --file string            target file/directory (default "transport-archive")
  -f, --force                  remove existing content
  -h, --help                   help for componentversions
      --lookup stringArray     repository name or spec for closure lookup fallback
  -O, --output string          output file for dry-run
  -S, --scheme string          schema version (default "v2")
  -s, --settings stringArray   settings file with variable settings (yaml)
      --templater string       templater to use (go, none, spiff, subst) (default "subst")
  -t, --type string            archive format (directory, tar, tgz) (default "directory")
  -v, --version string         default version for components
```

### Description


Add component versions specified by a description file to a Common Transport
Archive. This might be either a directory prepared to host component version
content or a tar/tgz file (see option --type).

If option <code>--create</code> is given, the archive is created first. An
additional option <code>--force</code> will recreate an empty archive if it
already exists.

If option <code>--complete</code> is given all component versions referenced by
the added one, will be added, also. Therefore, the <code>--lookup</code> is required
to specify an OCM repository to lookup the missing component versions. If
additionally the <code>-V</code> is given, the resources of those additional
components will be added by value.

The source, resource and reference list can be composed according to the commands
[ocm add sources](ocm_add_sources.md), [ocm add resources](ocm_add_resources.md), [ocm add references](ocm_add_references.md),
respectively.

The description file might contain:
- a single component as shown in the example
- a list of components under the key <code>components</code>
- a list of yaml documents with a single component or component list

The optional field <code>meta.configuredSchemaVersion</code> for a component
entry can be used to specify a dedicated serialization format to use for the
component descriptor. If given it overrides the <code>--schema</code> option
of the command. By default, v2 is used.


The <code>--type</code> option accepts a file format for the
target archive to use. The following formats are supported:
- directory
- tar
- tgz

The default format is <code>directory</code>.


If the option <code>--scheme</code> is given, the specified component descriptor format is used/generated.

The following schema versions are supported for explicit conversions:
  - <code>ocm.software/v3alpha1</code>
  - <code>v2</code> (default)


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


\
If a component lookup for building a reference closure is required
the <code>--lookup</code>  option can be used to specify a fallback
lookup repository. By default, the component versions are searched in
the repository holding the component version for which the closure is
determined. For *Component Archives* this is never possible, because
it only contains a single component version. Therefore, in this scenario
this option must always be specified to be able to follow component
references.


It the option <code>--copy-resources</code> is given, all referential
resources will potentially be localized, mapped to component version local
resources in the target repository. It the option <code>--copy-local-resources</code>
is given, instead, only resources with the relation <code>local</code> will be
transferred. This behaviour can be further influenced by specifying a transfer
script with the <code>script</code> option family.


### Examples




<pre>
$ ocm add componentversions &dash;&dash;file ctf &dash;&dash;version 1.0 components.yaml
</pre>


and a file <code>components.yaml</code>:


<pre>
name: ocm.software/demo/test
version: 1.0.0
provider:
  name: ocm.software
  labels:
    &dash; name: city
      value: Karlsruhe
labels:
  &dash; name: purpose
    value: test

resources:
  &dash; name: text
    type: PlainText
    input:
      type: file
      path: testdata
  &dash; name: data
    type: PlainText
    input:
      type: binary
      data: IXN0cmluZ2RhdGE=

</pre>


The resource <code>text</code> is taken from a file <code>testdata</code> located
next to the description file.


### SEE ALSO

##### Parents

* [ocm add](ocm_add.md)	 &mdash; Add resources or sources to a component archive
* [ocm](ocm.md)	 &mdash; Open Component Model command line client



##### Additional Links

* [<b>ocm add sources</b>](ocm_add_sources.md)	 &mdash; add source information to a component version
* [<b>ocm add resources</b>](ocm_add_resources.md)	 &mdash; add resources to a component version
* [<b>ocm add references</b>](ocm_add_references.md)	 &mdash; add aggregation information to a component version

