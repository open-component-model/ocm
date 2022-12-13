## plugin upload put &mdash; Upload Blob To External Repository

### Synopsis

```
plugin upload put [<flags>] <name> <repository specification> [<options>]
```

### Options

```
  -a, --artifactType string         artifact type of input blob
  -C, --credential <name>=<value>   dedicated credential value (default [])
  -c, --credentials YAML            credentials
  -h, --help                        help for put
  -H, --hint string                 reference hint for storing blob
  -m, --mediaType string            media type of input blob
```

### Description


Read the blob content from *stdin*, store the blob in the repository specified
by the given repository specification and return the access specification
(as JSON document string) usable to retrieve the blob, again, on * stdout*.
The uploader to use is specified by the first argument. This might only be
relevant, if the plugin supports multiple uploaders.


### SEE ALSO

##### Parents

* [plugin upload](plugin_upload.md)	 &mdash; upload specific operations
* [plugin](plugin.md)	 &mdash; OCM Plugin

