# Provide a GitOps template

This example shows how a GitOps template
can be manged via OCM tooling.
It is used to deliver deploy content and
a filesystem content template for a GitOps
environment. It additionally provides some
tiny tooling to manage upgrades on customer
side, including

- creating a new landscape based on a GitOps template
- provide new versions in the customer environment
- upgrade existing landscape projects (branches) based
  on the latest imported filesystem template

The automation in this example is done via github actions.
The ops part is left blank, any environment, for example
Flux can be used here.

The template projects is managed with *make*.
It can be used to create new versions (patch, minor, major)
and to build a component version that can later be
by the OCM tool set on the client side to import new
versions in a customer environment.
