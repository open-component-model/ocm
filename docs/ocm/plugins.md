# Extending the Library by Plugins

The library has several extension points,which can be used by a registration
mechanism to add further variants, like repository types, backend technologies,
access methods, blob downloaders and uploaders.

This requires Go coding, which is feasible for additional standard
implementations. Nevertheless, it is useful to provide a more dynamic 
way to enrich the functionality of the library and the OCM command line
tool.

This can be achieved by the experimental *plugin* concept. It allows 
to implement functionality in separate executables (the plugins) and
register them for any main program based on this library.

## Commands 

A plugin must provide a set of commands to implement the intended extension.
There is a set of commands for every extension point, which is supported
by the plugin.

The following extension points are generally possible to be
extended by a plugin:

- **access methods**: describe access to resource locations in foreign repositories.
- **blob uploaders**: export uploaded local blobs to foreign repositories.
- **blob downloaders**: transform downloaded resources to an applicable file system representation.


The documentation of the plugin commands can be found [here](../pluginreference/plugin.md).

## Implementation support

This library provides a command frame in package `pkg/contexts/ocm/plugin/ppi`.
It implements all the required command based on some interfaces, which must be
implemented by a plugin. These implementations are registered at a 
*Plugin*, which can then be passed to the standard implementation.

An example can be found in [`cmds/demoplugin`](https://github.com/open-component-model/ocm/tree/main/cmds/demoplugin).