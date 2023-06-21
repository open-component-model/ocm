## ocm logging &mdash; Configured Logging Keys

### Description


Logging can be configured as part of the ocm config file ([ocm configfile](ocm_configfile.md))
or by command line options of the [ocm](ocm.md) command. Details about
the YAML structure of a logging settings can be found on https://github.com/mandelsoft/logging.

The command line also supports some quick-config options for enabling log levels
for dedicated tags and realms (logging keys).

The following *tags* are used by the command line tool:
  - <code>blobhandler</code>: execution of blob handler used to upload resource blobs to an ocm repository.



The following *realms* are used by the command line tool:
  - <code>ocm</code>: general realm used for the ocm go library.
  - <code>ocm/accessmethod/ociartifact</code>: access method ociArtifact
  - <code>ocm/credentials/dockerconfig</code>: docker config handling as credential repository
  - <code>ocm/downloader</code>: Downloaders
  - <code>ocm/oci.ocireg</code>: OCI repository handling
  - <code>ocm/ocimapping</code>: OCM to OCI Registry Mapping
  - <code>ocm/plugins</code>: OCM plugin handling
  - <code>ocm/processing</code>: output processing chains
  - <code>ocm/refcnt</code>: reference counting
  - <code>ocm/toi</code>: TOI logging
  - <code>ocm/transfer</code>: OCM transfer handling



### Examples

```
type: logging.config.ocm.software
    contextType: attributes.context.ocm.software
    settings:
      defaultLevel: Info
      rules:
        - ...
```

### SEE ALSO

##### Parents

* [ocm](ocm.md)	 &mdash; Open Component Model command line client



##### Additional Links

* [<b>ocm configfile</b>](ocm_configfile.md)	 &mdash; configuration file

