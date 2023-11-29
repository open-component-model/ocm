# Signing Component Versions

This tour illustrates the basic functionality to
sign and verify signatures.

It covers two basic scenarios:
- [`sign`](01-basic-signing.go) Create, Sign, Transport and Verify a component version.
- [`repo`](02-using-context-settings.go) Using context settings to configure signing and verification in target repo.

You can just call the main program with some config file option (`--config <file>`) and the name of the scenario.
The config file should have the following content:

```yaml
targetRepository:
  type: CommonTransportFormat
  filePath: /tmp/example06.target.ctf
  fileFormat: directory
  accessMode: 2
ocmConfig: <your ocm config file>
```

The actual version of the example just works with the filesystem
target, because it is not possible to specify credentials for the
target repository in this simple config file. But, if you specific an [OCM config file](../04-working-with-config/README.md) you can
add more credential settings to make target repositories possible
requiring credentials.
