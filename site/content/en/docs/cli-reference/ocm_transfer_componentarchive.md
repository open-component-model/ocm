
---
title: ocm_transfer_componentarchive
url: /docs/cli-reference/ocm_transfer_componentarchive/
date: 2022-08-12T11:14:49+01:00
draft: false
images: []
menu:
  docs:
    parent: cli-reference
toc: true
---
## ocm transfer componentarchive &mdash; Transfer Component Archive To Some Component Repository

### Synopsis

```
ocm transfer componentarchive [<options>]  <source> <target>
```

### Options

```
  -h, --help          help for componentarchive
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

* [ocm transfer](ocm_transfer.md)	 - Transfer artefacts or components
* [ocm](ocm.md)	 - Open Component Model command line client

