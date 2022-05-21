## ocm componentarchive transfer &mdash; Transfer Component Archive To Some Component Repository

### Synopsis

```
ocm componentarchive transfer [<options>]  <source> <target>
```

### Options

```
  -h, --help          help for transfer
  -t, --type string   archive format (default "directory")
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

* [ocm componentarchive](ocm_componentarchive.md)	 - Commands acting on component archives
* [ocm](ocm.md)	 - Open Component Model command line client

