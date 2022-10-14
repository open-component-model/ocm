## ocm ocm componentarchive transfer &mdash; Transfer Component Archive To Some Component Repository

### Synopsis

```
ocm ocm componentarchive transfer [<options>]  <source> <target>
```

### Options

```
  -h, --help          help for transfer
  -t, --type string   archive format (directory, tar, tgz) (default "directory")
```

### Description


Transfer a component archive to some component repository. This might
be a CTF Archive or a regular repository.
If the type CTF is specified the target must already exist, if CTF flavor
is specified it will be created if it does not exist.

Besides those explicitly known types a complete repository spec might be configured,
either via inline argument or command configuration file and name.

The <code>--type</code> option accepts a file format for the
target archive to use. The following formats are supported:
- directory
- tar
- tgz
The default format is <code>directory</code>.


### SEE ALSO

##### Parents

* [ocm ocm componentarchive](ocm_ocm_componentarchive.md)	 &mdash; Commands acting on component archives
* [ocm ocm](ocm_ocm.md)	 &mdash; Dedicated command flavors for the Open Component Model
* [ocm](ocm.md)	 &mdash; Open Component Model command line client

