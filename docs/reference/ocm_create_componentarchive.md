
---
title: ocm_create_componentarchive
url: /docs/cli-reference/ocm_create_componentarchive/
date: 2022-08-12T11:14:49+01:00
draft: false
images: []
menu:
  docs:
    parent: cli-reference
toc: true
---
## ocm create componentarchive &mdash; Create New Component Archive

### Synopsis

```
ocm create componentarchive [<options>] <component> <version> <provider> <path> {--provider <label>=<value>} {<label>=<value>}
```

### Options

```
  -f, --force                  remove existing content
  -h, --help                   help for componentarchive
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

  - <code>ocm.gardener.cloud/v3alpha1</code>: 

  - <code>v2</code> (default): 



### SEE ALSO

##### Parents

* [ocm create](ocm_create.md)	 &mdash; Create transport or component archive
* [ocm](ocm.md)	 &mdash; Open Component Model command line client

