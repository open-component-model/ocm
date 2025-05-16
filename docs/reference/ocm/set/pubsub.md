---
title: "ocm set pubsub &mdash; Set The Pubsub Spec For An Ocm Repository"
linkTitle: "set pubsub"
url: "/docs/cli-reference/set/pubsub/"
sidebar:
  collapsed: true
---

### Synopsis

```bash
ocm set pubsub {<ocm repository>} [<pub/sub specification>]
```

#### Aliases

```text
pubsub, ps
```

### Options

```text
  -d, --delete   delete pub/sub configuration
  -h, --help     help for pubsub
```

### Description

A repository may be able to store a publish/subscribe specification
to propagate the creation or update of component versions.
If such an implementation is available this command can be used
to set the pub/sub specification for a repository.
If no specification is given an existing specification
will be removed for the given repository.
The specification
can be queried with the [ocm get pubsub](ocm_get_pubsub.md).
Types and specification formats are shown for the topic
[ocm ocm-pubsub](ocm_ocm-pubsub.md).

### SEE ALSO

#### Parents

* [ocm set](ocm_set.md)	 &mdash; Set information about OCM repositories
* [ocm](ocm.md)	 &mdash; Open Component Model command line client



##### Additional Links

* [<b>ocm get pubsub</b>](ocm_get_pubsub.md)	 &mdash; Get the pubsub spec for an ocm repository
* [<b>ocm ocm-pubsub</b>](ocm_ocm-pubsub.md)	 &mdash; List of all supported publish/subscribe implementations

