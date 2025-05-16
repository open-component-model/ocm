---
title: "ocm get config - Get Evaluated Config For Actual Command Call"
linkTitle: "get config"
url: "/docs/cli-reference/get/config/"
sidebar:
  collapsed: true
menu:
  docs:
    name: "get config"
---

### Synopsis

```bash
ocm get config <options>
```

#### Aliases

```text
config, cfg
```

### Options

```text
  -h, --help             help for config
  -O, --outfile string   output file or directory
  -o, --output string    output mode (JSON, json, yaml)
```

### Description

Evaluate the command line arguments and all explicitly
or implicitly used configuration files and provide
a single configuration object.


With the option <code>--output</code> the output mode can be selected.
The following modes are supported:
  - <code></code> (default)
  - <code>JSON</code>
  - <code>json</code>
  - <code>yaml</code>

### SEE ALSO

#### Parents

* [ocm get](ocm_get.md)	 &mdash; Get information about artifacts and components
* [ocm](ocm.md)	 &mdash; Open Component Model command line client

