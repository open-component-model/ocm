---
title: "ocm create transportarchive - Create New OCI/OCM Transport  Archive"
linkTitle: "create transportarchive"
url: "/docs/cli-reference/create/transportarchive/"
sidebar:
  collapsed: true
menu:
  docs:
    name: "create transportarchive"
---

### Synopsis

```bash
ocm create transportarchive [<options>] <path>
```

#### Aliases

```text
transportarchive, ctf
```

### Options

```text
  -f, --force         remove existing content
  -h, --help          help for transportarchive
  -t, --type string   archive format (directory, tar, tgz) (default "directory")
```

### Description

Create a new empty OCM/OCI transport archive. This might be either a directory prepared
to host artifact content or a tar/tgz file.

### SEE ALSO

#### Parents

* [ocm create](ocm_create.md)	 &mdash; Create transport or component archive
* [ocm](ocm.md)	 &mdash; Open Component Model command line client

