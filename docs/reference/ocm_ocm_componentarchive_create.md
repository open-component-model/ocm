## ocm ocm componentarchive create &mdash; Create New Component Archive

### Synopsis

```
ocm ocm componentarchive create [<options>] <component> <version> --provider <provider-name> {--provider <label>=<value>} {<label>=<value>}
```

### Options

```
  -F, --file string            target file/directory (default "component-archive")
  -f, --force                  remove existing content
  -h, --help                   help for create
  -p, --provider stringArray   provider attribute
  -S, --scheme string          schema version (default "v2")
  -t, --type string            archive format (directory, tar, tgz) (default "directory")
```

### Description


Create a new component archive. This might be either a directory prepared
to host component version content or a tar/tgz file (see option --type).

A provider must be specified, additional provider labels are optional.

The <code>--type</code> option accepts a file format for the
target archive to use. The following formats are supported:
- directory
- tar
- tgz
The default format is <code>directory</code>.

It the option <code>--scheme</code> is given, the given component descriptor format is used/generated.
The following schema versions are supported:

  - <code>ocm.gardener.cloud/v3alpha1</code>: 

  - <code>v2</code> (default): 



### SEE ALSO

##### Parents

* [ocm ocm componentarchive](ocm_ocm_componentarchive.md)	 &mdash; Commands acting on component archives
* [ocm ocm](ocm_ocm.md)	 &mdash; Dedicated command flavors for the Open Component Model
* [ocm](ocm.md)	 &mdash; Open Component Model command line client

