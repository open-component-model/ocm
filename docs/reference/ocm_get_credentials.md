## ocm get credentials &mdash; Get Credentials For A Dedicated Consumer Spec

### Synopsis

```
ocm get credentials {<consumer property>=<value>}
```

### Options

```
  -h, --help             help for credentials
  -m, --matcher string   matcher type override
```

### Description


Try to resolve a given consumer specification against the configured credential
settings and show the found credential attributes.

For the following usage contexts a standard identity matcher exists:
  - <code>OCIRegistry</code>: OCI registry credential matcher
  - <code>complete</code>: complete match of given pattern set
  - <code>hostpath</code>: Host and path based credential matcher
  - <code>partial</code>: complete match of given pattern ignoring additional attributes

The used matcher is derived from the consumer attribute <code>type</code>.
For all other consumer types a matcher matching all attributes will be used.
The usage of a dedicated matcher can be enforced by the option <code>--matcher</code>.


### SEE ALSO

##### Parents

* [ocm get](ocm_get.md)	 - Get information about artefacts and components
* [ocm](ocm.md)	 - Open Component Model command line client

