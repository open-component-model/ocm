{{config}}
# Working with Configurations

This tour illustrates the basic configuration management
included in the OCM library. The library provides
an extensible framework to bring together configuration settings
and configurable objects.

It covers five basic scenarios:
- [`basic`](01-basic-config-management.go) Basic configuration management illustrating the configuration of credentials.
- [`generic`](02-handle-arbitrary-config.go) Handling of arbitrary configuration.
- [`ocm`](03-using-ocm-config.go) Central configuration
- [`provide`](04-write-config-type.go) Providing new config object types
- [`consume`](05-write-config-consumer.go) Preparing objects to be configured by the config management

## Running the example

You can just call the main program with some config file option (`--config <file>`) and the name of the scenario.
The config file should have the following content:

```yaml
repository: ghcr.io/mandelsoft/ocm
username:
password:
```

Set your favorite OCI registry and don't forget to add the repository prefix for your OCM repository hosted in this registry.

## Walkthrough

### Basic configuration management

Similar to the other context areas, Configuration is handled by the configuration contexts.
Therefore, for the example, we just get the default configuration context.

```go
{{include}{../../04-working-with-config/01-basic-config-management.go}{default context}}
```

The configuration context handles configuration objects.
A configuration object is any object implementing
the `config.Config` interface. The task of a config object
is to apply configuration to some target object.

One such object is the configuration object for
credentials provided by the credentials context.

```go
{{include}{../../04-working-with-config/01-basic-config-management.go}{cred config}}
```

Here, we can configure credential settings:
credential repositories and consumer id mappings.
We do this by setting the credentials provided
by our config file for the consumer id used
by our configured OCI registry.

```go
{{include}{../../04-working-with-config/01-basic-config-management.go}{configure creds}}
```

(Credential) Configuration objects are typically serializable and deserializable.

```go
{{include}{../../04-working-with-config/01-basic-config-management.go}{marshal}}
```

Like all the other manifest based descriptions this format always includes
a type field, which can be used to deserialize a specification into
the appropriate object.
This can be done by the config context. It accepts YAML or JSON.

```go
{{include}{../../04-working-with-config/01-basic-config-management.go}{unmarshal}}
```

Regardless what variant is used (direct specification object or descriptor)
the config object can be added to a config context.

```go
{{include}{../../04-working-with-config/01-basic-config-management.go}{apply config}}
```

Every config object implements the
`ApplyTo(ctx config.Context, target interface{}) error` method.
It takes an object, which wants to be configured.
The config object then decides, whether it provides
settings for the given object and calls the appropriate
methods on this object (after a type cast).


Here, the code snippet from the apply method of the credential
config object ([.../pkg/contexts/credentials/config/type.go](../../../../../pkg/contexts/credentials/config/type.go)):

```go
{{include}{../../../../../pkg/contexts/credentials/config/type.go}{apply}}
        ...
```

This way the config mechanism reverts the configuration
request, it does not actively configure something, instead
an object, which wants to be configured calls the config
context to apply pending configs.
To do this the config context manages a queue of config objects
and applies them to an object to be configured.

If the credential context is asked now for credentials,
it asks the config context for pending config objects
and applies them.
Therefore, we now should get the configured credentials, here.

```go
{{include}{../../04-working-with-config/01-basic-config-management.go}{get credentials}}
```