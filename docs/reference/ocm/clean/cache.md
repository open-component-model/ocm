---
title: "ocm clean cache &mdash; Cleanup Oci Blob Cache"
linkTitle: "clean cache"
url: "/docs/cli-reference/clean/cache/"
sidebar:
  collapsed: true
---

### Synopsis

```bash
ocm clean cache [<options>]
```

### Options

```text
  -b, --before string   time since last usage
  -s, --dry-run         show size to be removed
  -h, --help            help for cache
```

### Description

Cleanup all blobs stored in oci blob cache (if given).
	
### Examples

```bash
$ ocm clean cache
```

### SEE ALSO

#### Parents

* [ocm clean](ocm_clean.md)	 &mdash; Cleanup/re-organize elements
* [ocm](ocm.md)	 &mdash; Open Component Model command line client

