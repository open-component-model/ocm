# Transporting Component Versions

This [tour](example.go) illustrates the basic support for
transporting content from one environment into another.


You can just call the main program with some config file option (`--config <file>`).
The config file should have the following content:

```yaml
repository: ghcr.io/mandelsoft/ocm
targetRepository:
  type: CommonTransportFormat
  filePath: /tmp/example05.target.ctf
  fileFormat: directory
  accessMode: 2
username:
password:
```

Any supported kind of target repository can be specified by using its
specification type. An OCI regisztry target would look like this:

```yaml
repository: ghcr.io/mandelsoft/ocm
username:
password:
targetRepository:
  type: OCIRegistry
  baseUrl: ghcr.io/mandelsoft/targetocm
ocmConfig: <config file>
```

The actual version of the example just works with the filesystem 
target, because it is not possible to specify credentials for the
target repository in this simple config file. But, if you specific an [OCM config file](../04-working-with-config/README.md) you can
add more credential settings to make target repositories possible
requiring credentials.