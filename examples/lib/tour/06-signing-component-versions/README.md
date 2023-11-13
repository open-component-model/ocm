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
```