## ocm ocm-pubsub &mdash; List Of All Supported Publish/Subscribe Implementations

### Description


OCM repositories can be configured to generate change events for
publish/subscribe systems, if there is a persistence provider
for the dedicated kind of OCM repository (for example OCI registry
based OCM repositories)


An OCM repository can be configured to propagate change events via a
publish/subscribe system, if there is a persistence provider for the dedicated
repository type. If available any know publish/subscribe system can
be configured with [ocm set pubsub](ocm_set_pubsub.md) and shown with
[ocm get pubsub](ocm_get_pubsub.md).. Hereby, the pub/sub system
is described by a typed specification.

The following list describes the supported publish/subscribe system types, their
specificaton versions and formats:

- PubSub type <code>compound</code>

  a pub/sub system forwarding events to described sub-level systems.

  The following versions are supported:
  - Version <code>v1</code>

    It is describe by the following field:

    - **<code>specifications</code>**  *list of pubsub specs*

      A list of nested sub-level specifications the events should be
      forwarded to.

There are persistence providers for the following repository types:
  - <code>OCIRegistry</code>


### SEE ALSO

##### Parents

* [ocm](ocm.md)	 &mdash; Open Component Model command line client



##### Additional Links

* [<b>ocm set pubsub</b>](ocm_set_pubsub.md)	 &mdash; Set the pubsub spec for an ocm repository
* [<b>ocm get pubsub</b>](ocm_get_pubsub.md)	 &mdash; Get the pubsub spec for an ocm repository

