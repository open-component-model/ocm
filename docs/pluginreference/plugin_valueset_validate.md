## plugin valueset validate &mdash; Validate Value Set

### Synopsis

```bash
plugin valueset validate <spec> [<options>]
```

### Options

```
  -h, --help   help for validate
```

### Description

This command accepts a value set as argument. It is used to
validate the specification and to provide some metadata for the given
specification.

This metadata has to be provided as JSON string on *stdout* and has the
following fields:

- **<code>description</code>** *string*

  A short textual description of the described value set.

### SEE ALSO

#### Parents

* [plugin valueset](plugin_valueset.md)	 &mdash; valueset operations
* [plugin](plugin.md)	 &mdash; OCM Plugin
