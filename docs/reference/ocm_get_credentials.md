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

For the following usage contexts with matchers and standard identity matchers exist:

  - <code>Buildcredentials.gardener.cloud</code>: Gardener config credential matcher
    
    It matches the <code>Buildcredentials.gardener.cloud</code> consumer type and additionally acts like 
    the <code>hostpath</code> type.

  - <code>OCIRegistry</code>: OCI registry credential matcher
    
    It matches the <code>OCIRegistry</code> consumer type and additionally acts like 
    the <code>hostpath</code> type.

  - <code>exact</code>: exact match of given pattern set

  - <code>hostpath</code>: Host and path based credential matcher
    
    This matcher works on the following properties:
    
    - *<code>hostname</code>* (required): the hostname of a server
    - *<code>port</code>* (optional): the port of a server
    - *<code>pathprefix</code>* (optional): a path prefix to match. The 
      element with the most matching path components is selected (separator is <code>/</code>).
    

  - <code>partial</code>: complete match of given pattern ignoring additional attributes


The used matcher is derived from the consumer attribute <code>type</code>.
For all other consumer types a matcher matching all attributes will be used.
The usage of a dedicated matcher can be enforced by the option <code>--matcher</code>.


### SEE ALSO

##### Parents

* [ocm get](ocm_get.md)	 &mdash; Get information about artefacts and components
* [ocm](ocm.md)	 &mdash; Open Component Model command line client

