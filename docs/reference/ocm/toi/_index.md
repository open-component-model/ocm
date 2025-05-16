---
title: "ocm toi - Dedicated Command Flavors For The TOI Layer"
linkTitle: "toi"
url: "/docs/cli-reference/toi/"
sidebar:
  collapsed: true
menu:
  docs:
    name: "toi"
---

### Synopsis

```bash
ocm toi [<options>] <sub command> ...
```

### Options

```text
  -h, --help   help for toi
```

### Description

TOI is an abbreviation for Tiny OCM Installation. It is a simple
application framework on top of the Open Component Model, that can
be used to describe image based installation executors and installation
packages (see topic [ocm toi-bootstrapping](ocm_toi-bootstrapping.md) in form of resources
with a dedicated type. All involved resources are hereby taken from a component
version of the Open Component Model, which supports all the OCM features, like
transportation.

The framework consists of a generic bootstrap command
([ocm bootstrap package](ocm_bootstrap_package.md)) and an arbitrary set of image
based executors, that are executed in containers and fed with the required
installation data by th generic command.

### SEE ALSO

#### Parents

* [ocm](ocm.md)	 &mdash; Open Component Model command line client


##### Sub Commands

* ocm toi <b>bootstrap</b>	 &mdash; bootstrap components
* ocm toi <b>configuration</b>	 &mdash; TOI Commands acting on config
* ocm toi <b>describe</b>	 &mdash; describe packages
* ocm toi <b>package</b>	 &mdash; TOI Commands acting on components



##### Additional Help Topics

* [ocm toi-bootstrapping](ocm_toi-bootstrapping.md)	 &mdash; Tiny OCM Installer based on component versions
* [ocm <b>ocm-references</b>](ocm_ocm-references.md)	 &mdash; notation for OCM references


##### Additional Links

* [<b>ocm bootstrap package</b>](ocm_bootstrap_package.md)	 &mdash; bootstrap component version

