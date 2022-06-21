## ocm componentarchive create &mdash; Create New Component Archive

### Synopsis

```
ocm componentarchive create [<options>] <component> <version> <provider> <path> {--provider <label>=<value>} {<label>=<value>}
```

### Options

```
  -f, --force                  remove existing content
  -h, --help                   help for create
  -p, --provider stringArray   provider attribute
  -S, --scheme string          schema version (default "v2")
  -t, --type string            archive format (default "directory")
```

### Description


Create a new component archive. This might be either a directory prepared
to host component version content or a tar/tgz file.
With option <code>-S</code> it is possible to specify the intended scheme version.
The following versions are currently supported:

  - <code>ocm.gardener.cloud/v3</code>
  - <code>v2</code> (default)


### SEE ALSO

##### Parents

* [ocm componentarchive](ocm_componentarchive.md)	 - Commands acting on component archives
* [ocm](ocm.md)	 - Open Component Model command line client

