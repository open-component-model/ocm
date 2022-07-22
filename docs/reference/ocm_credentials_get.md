## ocm credentials get &mdash; Get Credentials For A Dedicated Consumer Spec

### Synopsis

```
ocm credentials get {<consumer property>=<value>}
```

### Options

```
  -h, --help             help for get
  -m, --matcher string   matcher type override
```

### Description


Try to resolve a given consumer specification against the configured credential
settings and show the found credential attributes.

For the following usage contexts with matchers and standard identity matchers exist:

  - <code>OCIRegistry</code>: OCI registry credential matcher
  - <code>exact</code>: exact match of given pattern set
  - <code>partial</code>: complete match of given pattern ignoring additional attributes

The used matcher is derived from the consumer attribute <code>type</code>.
For all other consumer types a matcher matching all attributes will be used.
The usage of a dedicated matcher can be enforced by the option <code>--matcher</code>.


### SEE ALSO

##### Parents

* [ocm credentials](ocm_credentials.md)	 - Commands acting on credentials
* [ocm](ocm.md)	 - Open Component Model command line client

