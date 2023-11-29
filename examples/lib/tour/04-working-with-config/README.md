# Working with Configurations

This tour illustrates the basic configuration management
included in the OCM library. The library provides
an extensible framework to bring together configuration settings
and configurable objects.

It covers five basic scenarios:
- [`basic`](01-basic-config-management.go) Basic configuration management illustarting the configuration of credentials.
- [`generic`](02-handle-arbitrary-config.go) Handling of arbitrary configuration.
- [`ocm`](03-using-ocm-config.go) Central configuration
- [`provide`](04-write-config-type.go) Providing new config object types
- [`consume`](05-write-config-consumer.go) Preparing objects to be configured by the config management


You can just call the main program with some config file option (`--config <file>`) and the name of the scenario.
The config file should have the following content:

```yaml
repository: ghcr.io/mandelsoft/ocm
username:
password:
```