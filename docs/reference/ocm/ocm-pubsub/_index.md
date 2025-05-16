---
title: "ocm ocm-pubsub - List Of All Supported Publish/Subscribe Implementations"
linkTitle: "ocm-pubsub"
url: "/docs/cli-reference/ocm-pubsub/"
sidebar:
  collapsed: true
menu:
  docs:
    name: "ocm-pubsub"
---

### Description

An OCM repository can be configured to propagate change events via a
publish/subscribe system, if there is a persistence provider for the dedicated
repository type. If available any known publish/subscribe system can
be configured with [ocm set pubsub](ocm_set_pubsub.md) and shown with
[ocm get pubsub](ocm_get_pubsub.md). Hereby, the pub/sub system
is described by a typed specification.


The following list describes the supported publish/subscribe system types, their
specification versions, and formats:

- PubSub type <code>compound</code>

  A pub/sub system forwarding events to described sub-level systems.

  The following versions are supported:
  - Version <code>v1</code>

    It is described by the following field:

    - **<code>specifications</code>**  *list of pubsub specs*

      A list of nested sub-level specifications the events should be
      forwarded to.


- PubSub type <code>redis</code>

  a redis pubsub system.

  The following versions are supported:
  - Version <code>v1</code>

    It is describe by the following field:

    - **<code>serverAddr</code>**  *Address of redis server*
    - **<code>channel</code>**  *pubsub channel*
    - **<code>database</code>**  *database number*

      Publishing using the redis pubsub API. For every change a string message
      with the format <component>:<version> is published. If multiple repositories
      should be used, each repository should be configured with a different
      channel.

There are persistence providers for the following repository types:
  - <code>OCIRegistry</code>

### SEE ALSO

#### Parents

* [ocm](ocm.md)	 &mdash; Open Component Model command line client



##### Additional Links

* [<b>ocm set pubsub</b>](ocm_set_pubsub.md)	 &mdash; Set the pubsub spec for an ocm repository
* [<b>ocm get pubsub</b>](ocm_get_pubsub.md)	 &mdash; Get the pubsub spec for an ocm repository

