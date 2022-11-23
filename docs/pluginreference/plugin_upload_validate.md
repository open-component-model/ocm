## plugin upload validate &mdash; Validate Upload Specification

### Synopsis

```
plugin upload validate [<flags>] <name> <spec> [<options>]
```

### Options

```
  -a, --artefactType string   artefact type of input blob
  -h, --help                  help for validate
  -m, --mediaType string      media type of input blob
```

### Description


This command accepts a target specification as argument. It is used to
validate the specification for the specified upoader and to provide some
metadata for the given specification.

This metadata has to be provided as JSON document string on *stdout* and has the
following fields:

- **<code>consumerId</code>** *map[string]string*

  The consumer id used to determine optional credentials for the
  underlying repository. If specified, at least the <code>type</code> field must
  be set.


### SEE ALSO

##### Parents

* [plugin upload](plugin_upload.md)	 &mdash; upload specific operations
* [plugin](plugin.md)	 &mdash; OCM Plugin

