## plugin download &mdash; Download Blob Into Filesystem

### Synopsis

```
plugin download [<flags>] <name> <filepath> [<options>]
```

### Options

```
  -a, --artifactType string   artifact type of input blob
  -c, --config string         registration config
  -h, --help                  help for download
  -m, --mediaType string      media type of input blob
```

### Description


This command accepts a target filepath as argument. It is used as base name
to store the downloaded content. The blob content is provided on the
*stdin*. The first argument specified the downloader to use for the operation.

The task of this command is to transform the content of the provided 
blob into a filesystem structure applicable to the type specific tools working
with content of the given artifact type.


### SEE ALSO

##### Parents

* [plugin](plugin.md)	 &mdash; OCM Plugin

