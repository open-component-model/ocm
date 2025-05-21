# OCM Plugin framework

THIS IS A TEST EDIT.

The OCM Plugin framework now supports two features to
extend the CLI with new (OCM related) commands:

- definition of configuration types (consumed by the plugin)
- definition of CLI commands (for the OCM CLI)

Additionally, it is possible to consume logging configuration from the OCM CLI for all
plugin feature commands.

Examples see coding in `cmds/cliplugin`

## Config Types

Config types are just registered at the Plugin Object;

```go
    p := ppi.NewPlugin("cliplugin", version.Get().String())
        ...
    p.RegisterConfigType(configType)
```

The argument is just the config type as registered at the ocm library, for example:

```go
const ConfigType = "rhabarber.config.acme.org"

type Config struct {
    runtime.ObjectVersionedType `json:",inline"`
    ...
}

func (a *Config) ApplyTo(ctx cfgcpi.Context, target interface{}) error {
        ...
}

func init() {
    configType = cfgcpi.NewConfigType[*Config](ConfigType, usage)
    cfgcpi.RegisterConfigType(configType)
}
```

## CLI Commands

CLI commands are provided by the registration interface `ppi.Command`. It
provides some command metadata and a `cobra.Command` object.

Commands are then registered at the plugin object with:

```go
    p.RegisterCommand(cmd)
```

The plugin programming interface supports the generation of an extension command directly from a
`cobra.command` object using the method `NewCLICommand` from the `ppi.clicmd` package.
It takes some options to specify the command embedding and extracts the other command attributes
directly from the preconfigured cobra command.

Otherwise, the `ppi.Command` interface  can be implemented without requiring a cobra command.

A sample code could look like this:

```go
    cmd, err := clicmd.NewCLICommand(NewCommand(), clicmd.WithCLIConfig(), clicmd.WithVerb("check"))
    if err != nil {
        os.Exit(1)
    }
    p.RegisterCommand(cmd)
```

with coding for the cobra command similar to

```go
type command struct {
    date string
}

func NewCommand() *cobra.Command {
    cmd := &command{}
    c := &cobra.Command{
        Use:   Name + " <options>",
        Short: "determine whether we are in rhubarb season",
        Long:  "The rhubarb season is between march and april.",
        RunE:  cmd.Run,
    }

    c.Flags().StringVarP(&cmd.date, "date", "d", "", "the date to ask for (MM/DD)")
    return c
}

func (c *command) Run(cmd *cobra.Command, args []string) error {
   ...
}
```

If the code wants to use the config framework, for example to

- use the OCM library again
- access credentials
- get configured with declared config types

the appropriate command feature must be set.
For the cobra support this is implemented by the option `WithCLIConfig`.
If set to true, the OCM CLI configuration is available for the config context used in the
CLI code.

The command can be a top-level command or attached to a dedicated verb (and optionally a realm like `ocm`or `oci`).
For the cobra support this can be requested by the option `WithVerb(...)`.

If the config framework is used just add the following anonymous import
for an automated configuration:

```go
import (
        // enable mandelsoft plugin logging configuration.
    _ "ocm.software/ocm/pkg/contexts/ocm/plugin/ppi/config"
)
```

The plugin code is then configured with the configuration of the OCM CLI and the config  framework
can be used.
If the configuration should be handled by explicit plugin code a handler can be registered with

```go
func init() {
    command.RegisterCommandConfigHandler(yourHandler)
}
```

It gets a config yaml according to the config objects used by the OCM library.

## Logging

To get the logging configuration from the OCM CLI the plugin has be configured with

```go
    p.ForwardLogging()
```

If the standard mandelsoft logging from the OCM library is used the configuration can
be implemented directly with an anonymous import of

```go
import (
        // enable mandelsoft plugin logging configuration.
    _ "ocm.software/ocm/pkg/contexts/ocm/plugin/ppi/logging"
)
```

The plugin code is then configured with the logging configuration of the OCM CLI and the mandelsoft logging frame work
can be used.
If the logging configuration should be handled by explicit plugin code a handler can be registered with

```go
func init() {
    cmds.RegisterLoggingConfigHandler(yourHandler)
}
```

It gets a logging configuration yaml according to the logging config used by the OCM library (`github.com/mandelsoft/logging/config`).

## Using Plugin command extensions from the OCM library

The plugin command extensions can also be called without the OCM CLI directly from the OCM library.
Therefore the plugin objects provided by the library can be used.

Logging information and config information must explicitly be configured to be passed to the
plugin.

Therefore the context attribute `clicfgattr.Get(ctx)` is used. It can be set via `clicfgattr.Set(...)`.
The logging configuration is extracted from the configured configuration object with target type `*logging.LoggingConfiguration`.

If the code uses an OCM context configured with a `(ocm)utils.ConfigureXXX` function, the cli config attribute is set accordingly.
