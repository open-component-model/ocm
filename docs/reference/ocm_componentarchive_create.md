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

The <code>--type</code> option accepts a file format for the
target archive to use. The following formats are supported:
- directory
- tar
- tgz
The default format is <code>directory</code>.

It the option <code>--scheme</code> is given, the given component descriptor format is used/generated.
The following schema versions are supported:

  - <code>ocm.gardener.cloud/v3alpha1</code>
  - <code>v2</code> (default)


### SEE ALSO

##### Parents

* [ocm componentarchive](ocm_componentarchive.md)	 - Commands acting on component archives
* [ocm](ocm.md)	 - Open Component Model command line client

