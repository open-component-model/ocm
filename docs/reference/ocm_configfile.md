## ocm configfile &mdash; Configuration File

### Description


The command line client supports configuring by a given configuration file.
If existent by default the file <code>$HOME/.ocmconfig</code> will be read.
Using the option <code>--config</code> an alternative file can be specified.

The file format is yaml. It uses the same type mechanism used for all
kinds of typed specification in the ocm area. The file must have the type of
a configuration specification. Instead, the command line client supports
a generic configuration specification able to host a list of arbitrary configuration
specifications. The type for this spec is <code>generic.config.ocm.gardener.cloud/v1</code>.


### Examples

```

type: generic.config.ocm.gardener.cloud/v1
configurations:
  - type: credentials.config.ocm.gardener.cloud
    repositories:
      - repository:
          type: DockerConfig/v1
          dockerConfigFile: "~/.docker/config.json"
          propagateConsumerIdentity: true
   - type: attributes.config.ocm.gardener.cloud
     attributes:  # map of attribute settings
       compat: true
#  - type: scripts.ocm.config.ocm.gardener.cloud
#    scripts:
#      "default":
#         script:
#           process: true

```

### SEE ALSO

##### Parents

* [ocm](ocm.md)	 - Open Component Model command line client

