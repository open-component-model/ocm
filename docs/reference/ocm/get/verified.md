---
title: "get verified"
url: "/docs/cli-reference/get/verified/"
sidebar:
  collapsed: true
---

## ocm get verified &mdash; Get Verified Component Versions

### Synopsis

```bash
ocm get verified [<options>] {<component / version}
```

### Options

```text
  -h, --help               help for verified
  -o, --output string      output mode (JSON, json, wide, yaml)
  -s, --sort stringArray   sort fields
      --verified string    verified file (default "~/.ocm/verified")
```

### Description

Get lists remembered verified component versions.


With the option <code>--output</code> the output mode can be selected.
The following modes are supported:
  - <code></code> (default)
  - <code>JSON</code>
  - <code>json</code>
  - <code>wide</code>
  - <code>yaml</code>

### Examples

```text
$ ocm get verified
$ ocm get verified -f verified.yaml acme.org/component -o yaml
```

### SEE ALSO

#### Parents

* [ocm get](ocm_get.md)	 &mdash; Get information about artifacts and components
* [ocm](ocm.md)	 &mdash; Open Component Model command line client

