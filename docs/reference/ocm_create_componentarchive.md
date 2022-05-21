## ocm create componentarchive &mdash; Create New Component Archive

### Synopsis

```
ocm create componentarchive [<options>] <component> <version> <provider> <path> {<label>=<value>}
```

### Options

```
  -f, --force         remove existing content
  -h, --help          help for componentarchive
  -t, --type string   archive format (default "directory")
```

### Description


Create a new component archive. This might be either a directory prepared
to host component version content or a tar/tgz file.


### SEE ALSO

##### Parents

* [ocm create](ocm_create.md)	 - Create transport or component archive
* [ocm](ocm.md)	 - Open Component Model command line client

