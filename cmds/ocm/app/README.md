# Open Component Model Command Line Tool

## Config File

The command line tool can be configured by a configuration file. If not
specified on the command line, the file `~/.ocmconfig` is read.

The configuration file is a yaml file following format by the
[configuration context](../../../api/config/README.md).

It consists of list of configuration specifications according to
the registered configurations types provided by the used library.
Every entry must provide an appropriate type field specifying
the dedicated configuration format.

The basic layout looks as follows:

```yaml
type: generic.config.ocm.software/v1
configurations:
  - type: credentials.config.ocm.software
    repositories:
      - repository:
          type: DockerConfig/v1
          dockerConfigFile: "~/.docker/config.json"
          propagateConsumerIdentity: true
```

The example above just lists a configuration specification
supported by the credentials context, which configures
the docker configuration file as credential repository to use.
Additionally it is configured to assign the contained credentials
to their OCI repositories managed by the OCI context.