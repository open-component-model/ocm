# The OCM Library Plugin Concept

The ocm library supports a plugin mechanism to provide further variants
for OCM and library extension points without the need of extending and recompiling
the OCM CLI (and other applications using this library).

This plugin concept is not part of the OCM specification, because it is
a feature of this library implementation.

The following extension points are supported:

- Access methods
- Uploaders
- Downloaders
- Actions
- Value sets (for example for routing slip entries)
- Config types
- Value Merge Handler (for example for label values in delta transports)
- CLI Commands (for OCM CLI)
- (transfer handlers)
- (signing tools)
- (input types for component version composition)

## Plugin Technology

A plugin is a simple executable, which might be written in any program language.
The plugin has to provide a set of CLI commands for every extension point it provides new variations for.

The data transfer between the library and the plugin is done via

- command options (for inbound information)
- standard input (for potentially large inbound content (streaming))
- standard output (for structured outbound information)
- standard output (for unstructured data (streaming)

The commands and their interface are described in the [plugin reference](../../../docs/pluginreference/plugin.md).

Every plugin must provide the [`info`](../../../docs/pluginreference/plugin_info.md) command. It has to provide information
about the supported features as JSOn document on standard output.
The structure of this document is described by the descriptor type in package [`api/ocm/plugin/descriptor`](descriptor/descriptor.go)

Plugin are searched in a plugin folder (typically `.ocm/plugins`), This default location can be changed by the `plugincachedir` attribute.

## Plugin-related CLI Commands

The OCM CLI provides commands to

- [install](../../../docs/reference/ocm_install_plugins.md)
- [update](../../../docs/reference/ocm_install_plugins.md)
- [list](../../../docs/reference/ocm_get_plugins.md)
- [examine](../../../docs/reference/ocm_describe_plugins.md)

plugins.

Plugins can either be installed manually, just by copying the plugin executable to the plugin directory, or by using the CLI commands. They uye OCM component versions as installation source. Plugins must have the artifact type `ocmPlugin` and follow the rules for providing multi-platform executables by using separate
resources with the same name by different platform attributes as extended identity. (see Go platform and architecture names).

The commands extract the correct variant for the platform the command is running.
If the given reference does not include a resource identity, the first resource with the correct artifact type is used.
plugins.

## Support for writing Plugins

To write an OCM plugin in Go, the provided [support library](ppi) can be used.
It provides a complete set of commands for all extension points and a main
function to run the plugin.

It can be used by a `main` function to run the plugin:

```Go
package main

import (
	"os"

	"ocm.software/ocm/api/ocm/plugin/ppi"
	"ocm.software/ocm/api/ocm/plugin/ppi/cmds"
	"ocm.software/ocm/api/version"
	"ocm.software/ocm/cmds/demoplugin/accessmethods"
	"ocm.software/ocm/cmds/demoplugin/config"
	"ocm.software/ocm/cmds/demoplugin/uploaders"
	"ocm.software/ocm/cmds/demoplugin/valuesets"
)

func main() {
	p := ppi.NewPlugin("demo", version.Get().String())

	p.SetShort("demo plugin")
	p.SetLong("plugin providing access to temp files and a check routing slip entry.")
	p.SetConfigParser(config.GetConfig)

	p.RegisterAccessMethod(accessmethods.New())
	u := uploaders.New()
	p.RegisterUploader("testArtifact", "", u)
	p.RegisterValueSet(valuesets.New())
	err := cmds.NewPluginCommand(p).Execute(os.Args[1:])
	if err != nil {
		os.Exit(1)
	}
}
```

The library provides an interface type [Plugin](ppi/interface.go) and a standard [implementation](ppi/plugin.go) for this interface.

It is used to configure the desired extension point implementations. Every extension point provides an extension interface and supporting types.
By implementing and registering such an interface the plugin object gets
informed about the implemented extension point variants and how they are implemented.

This information is then used by the standard implementation of the `info` command
to provide an appropriate plugin descriptor.

The support provides a complete command set. Therefore, the extension point implementation have never to deal with the command line interface, everything is described by the provided extension point interfaces.

The implemented standard commands have access to the plugin object and therefore
determine whether there is an implementation for an extension point variant requested via the command execution.

If there is an implementation, the CLI interface is just mapped to calls to interface functions of the particular extension point functionality.

The standard implementation od the commands can be found below the package
[`api/plugin/ppi/cmds](ppi/cmds), structured by the extension point name and its
interface operation. (for example [upload/put](ppi/cmds/upload/put/cmd.go))

## How plugins are used in the library

The library includes the counterpart of the plugin-side support. It is found directly in package [`api/ocm/plugin](.). It shared the descriptor package for the plugin descriptor and provides a separate [Plugin](plugin.go) object type.

Such an object provides methods for all extension point functions, which can the appropriate CLI command for its plugin executable. Again, it shields the CLI interface by providing an appropriate Go interface.

To enable the usage of plugins in an [api/ocm/Context](../internal/context.go),
the plugins must be registered at this context by calling [api/ocm/plugin/registration/RegisterExtensions](registration/registration.go).
It used the plugin cache die attribute to setup a plugin cache by reading all the plugin descriptors of the configured plugins and providing appropriate plugin objects.

For extension points requiring a static registration (like access method) it uses
the information from the descriptors to determine the provided variants and registers appropriate proxy type implementations at the context.

For extension points supporting an on-demand registration it use
the registration handler registration feature of those extension points
to register an appropriate registration handler using the namespace prefix `plugin`. This handler evaluate sub names to determine the plugin name and extension name in this plugin to provide an extension handler implementation proxy to for ward the handler functionality to the appropriate plugin.

### Extension Proxies

Regardless, whether it is a static type proxy or handler proxy, the proxy implementation keeps track of it plugin and extension name and maps
the extension point functionality to the Go interface of the plugin representation, which the calls the appropriate CLI command of the plugin executable.

The proxy implementation are found in `plugin` packages below the extension point package (for example the [access method plugin proxy](../extensions/accessmethods/plugin)).

The registration handler registries are typically found in the extension point packages in (for example)) [registrations.go](../extensions/download/registration.go). The generale registration handler registration handling is implemented in package [api/utils/registrations](../../utils/registrations).
It is called registration handler registry, because it is a registry which provides a namespace to name the types of handlers. It is possible to register registration handlers for a sub namespace, which then handle the registration (and creation) of handlers for a particular extension point (for example download handler).

The `command` extension point is implemented by the CLI package [cmds/ocm/commands/ocmcmds/plugins](../../../cmds/ocm/commands/ocmcmds/plugins).
Command extensions can be registered for any verb and object type name (even those not yet existing) and are automatically added to the CLI's command tree.
