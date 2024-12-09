## plugin &mdash; OCM Plugin

### Synopsis

```bash
plugin <subcommand> <options> <args>
```

### Options

```text
  -c, --config YAML       plugin configuration
  -h, --help              help for plugin
      --log-config YAML   ocm logging configuration
```

### Description

The OCM library has several extension points, which can be used by a registration
mechanism to add further variants, like repository types, backend technologies,
access methods, blob downloaders and uploaders.

This requires Go coding, which is feasible for additional standard
implementations. Nevertheless, it is useful to provide a more dynamic
way to enrich the functionality of the library and the OCM command line
tool.

This can be achieved by the experimental *plugin* concept. It allows
to implement functionality in separate executables (the plugins) and
register them for any main program based on this library.

A plugin may contribute to the following extension points:
- **access methods**: describe access to resource locations in foreign repositories.
- **blob uploaders**: export uploaded local blobs to foreign repositories.
- **blob downloaders**: transform downloaded resources to an applicable file system representation.

## Commands

A plugin must provide a set of commands to implement the intended extension.

The library allows to configure settings for a plugin, this configuration
is optionally passed to all commands as JSON argument using option <code>-c</code>.

Errors have to be reported on *stderr* as JSON document with the fields:

- **<code>error</code>** *string*

  The error message provided by a command.

Any plugin, regardless of its functionality has to provide an [plugin info](plugin_info.md),
which prints JSON document containing a plugin descriptor that describes the
capabilities of the plugin.

### SEE ALSO



##### Sub Commands

* [plugin <b>accessmethod</b>](plugin_accessmethod.md)	 &mdash; access method operations
* [plugin <b>action</b>](plugin_action.md)	 &mdash; action operations
* [plugin <b>describe</b>](plugin_describe.md)	 &mdash; describe plugin
* [plugin <b>download</b>](plugin_download.md)	 &mdash; download blob into filesystem
* [plugin <b>info</b>](plugin_info.md)	 &mdash; show plugin descriptor
* [plugin <b>transferhandler</b>](plugin_transferhandler.md)	 &mdash; decide on a question related to a component version transport
* [plugin <b>upload</b>](plugin_upload.md)	 &mdash; upload specific operations
* [plugin <b>valuemergehandler</b>](plugin_valuemergehandler.md)	 &mdash; value merge handler operations
* [plugin <b>valueset</b>](plugin_valueset.md)	 &mdash; valueset operations



##### Additional Help Topics

* [plugin <b>command</b>](plugin_command.md)	 &mdash; CLI command extensions
* [plugin <b>descriptor</b>](plugin_descriptor.md)	 &mdash; Plugin Descriptor Format Description
