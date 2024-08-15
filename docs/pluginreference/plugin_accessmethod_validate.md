## plugin accessmethod validate &mdash; Validate Access Specification

### Synopsis

```bash
plugin accessmethod validate <spec> [<options>]
```

### Options

```text
  -h, --help   help for validate
```

### Description

This command accepts an access specification as argument. It is used to
validate the specification and to provide some metadata for the given
specification.

This metadata has to be provided as JSON string on *stdout* and has the
following fields:

- **<code>mediaType</code>** *string*

  The media type of the artifact described by the specification. It may be part
  of the specification or implicitly determined by the access method.

- **<code>description</code>** *string*

  A short textual description of the described location.

- **<code>hint</code>** *string*

  A name hint of the described location used to reconstruct a useful
  name for local blobs uploaded to a dedicated repository technology.

- **<code>consumerId</code>** *map[string]string*

  The consumer id used to determine optional credentials for the
  underlying repository. If specified, at least the <code>type</code> field must be set.

### SEE ALSO

#### Parents

* [plugin accessmethod](plugin_accessmethod.md)	 &mdash; access method operations
* [plugin](plugin.md)	 &mdash; OCM Plugin

