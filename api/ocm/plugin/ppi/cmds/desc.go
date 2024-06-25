package cmds

func Description(name string) string {
	return `
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

Any plugin, regardless of its functionality has to provide an <CMD>` + name + ` info</CMD>,
which prints JSON document containing a plugin descriptor that describes the 
apabilities of the plugin.
`
}
