---
title: "get plugins"
menu:
  docs:
    parent: get
---
## ocm get plugins &mdash; Get Plugins

### Synopsis

```bash
ocm get plugins [<options>] {<plugin name>}
```

#### Aliases

```text
plugins, plugin, p
```

### Options

```text
  -h, --help               help for plugins
  -o, --output string      output mode (JSON, json, wide, yaml)
  -s, --sort stringArray   sort fields
```

### Description

Get lists information for all plugins specified, if no plugin is specified
all registered ones are listed.


With the option <code>--output</code> the output mode can be selected.
The following modes are supported:
  - <code></code> (default)
  - <code>JSON</code>
  - <code>json</code>
  - <code>wide</code>
  - <code>yaml</code>

### Examples

```bash
$ ocm get plugins
$ ocm get plugins demo -o yaml
```

### SEE ALSO

#### Parents

* [ocm get](ocm_get.md)	 &mdash; Get information about artifacts and components
* [ocm](ocm.md)	 &mdash; Open Component Model command line client

