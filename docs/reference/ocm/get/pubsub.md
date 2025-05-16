---
title: "ocm get pubsub &mdash; Get The Pubsub Spec For An Ocm Repository"
linkTitle: "get pubsub"
url: "/docs/cli-reference/get/pubsub/"
sidebar:
  collapsed: true
---

### Synopsis

```bash
ocm get pubsub {<ocm repository>}
```

#### Aliases

```text
pubsub, ps
```

### Options

```text
  -h, --help               help for pubsub
  -o, --output string      output mode (JSON, json, yaml)
  -s, --sort stringArray   sort fields
```

### Description

A repository may be able to store a publish/subscribe specification
to propagate the creation or update of component versions.
If such an implementation is available and a specification is
assigned to the repository, it is shown. The specification
can be set with the [ocm set pubsub](ocm_set_pubsub.md).


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



##### Additional Links

* [<b>ocm set pubsub</b>](ocm_set_pubsub.md)	 &mdash; Set the pubsub spec for an ocm repository

