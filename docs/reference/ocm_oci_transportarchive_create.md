
---
title: ocm_oci_transportarchive_create
url: /docs/cli-reference/ocm_oci_transportarchive_create/
date: 2022-08-12T11:14:49+01:00
draft: false
images: []
menu:
  docs:
    parent: cli-reference
toc: true
---
## ocm oci transportarchive create &mdash; Create New OCI/OCM Transport  Archive

### Synopsis

```
ocm oci transportarchive create [<options>] <path>
```

### Options

```
  -f, --force         remove existing content
  -h, --help          help for create
  -t, --type string   archive format (default "directory")
```

### Description


Create a new empty OCM/OCI transport archive. This might be either a directory prepared
to host artefact content or a tar/tgz file.


### SEE ALSO

##### Parents

* [ocm oci transportarchive](ocm_oci_transportarchive.md)	 &mdash; Commands acting on OCI view of a Common Transport Archive
* [ocm oci](ocm_oci.md)	 &mdash; Dedicated command flavors for the OCI layer
* [ocm](ocm.md)	 &mdash; Open Component Model command line client

