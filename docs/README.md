# Open Component Model (OCM)

The goal of the Open Component Model is to describe a machine-readable format 
for software-bill-of-deliveries (SBOD) with the focus on the delivery
artifacts. Primarily, it does not deal with the content or packages those
artifacts are composed of. This kind of meta information can be separately
attached to described artifacts by labels, separate resources or even by separate
components.

It is a completely technology-agnostic model to describe artifacts and
the technical access to their content. Technology-agnostic means:

- it can describe any artifact regardless of its technology
- the artifacts can basically be stored using any storage backend technology or
  repository
- the model information can be stored using any storage backend technology or
  repository

The only constraint is, that there must be
- an implementation for accessing artifacts in the desired repository technology 
  and map them to a blob format
- and a specification for a [mapping scheme](ocm/interoperability.md) describing how
  to map the elements
  of the component model to the supported elements of the backend technology
- and an [implementation](ocm/operations.md) of all the mapping schemes for the
  storage scenarios used in a dedicated environment.

The model uses a globally unique naming scheme for software [components](ocm/model.md#components).
Components are versioned. Every component version describes
a set of formally typed delivery artifacts (like OCI images). Those artifacts
get assigned unique identities in the context of the component version.

Those artifact definitions may carry an additional arbitrary attribution, and 
they provide a formal specification of the access method, which can be used
to access the technical content of the artifact in the actual evaluation
context of a component version.

The description model allows to transport content from one repository 
landscape (hosting the real technical artifacts and the component version
descriptions) into other, especially private,
environments without losing the validity of the access information. In any
environment the actual description of the component version carries valid
environment-specific information for the artifact location.

A transport tool can use the bill of material to determine the set of
artifacts that have to be transferred and use the access information to access
the technical content of the artifacts in the source environment. They will then
be copied into a repository landscape on the target side. In the target
environment, the model description will be stored, also, after it has been
adapted to reflect the new location of the described artifacts.

Using provided implementations for the used access types, this can be done 
in completely generic manner if there is a common interface and a discovery
mechanism for the implementation of the access methods, based on the type
information stored along with the artifact description in the component version.

This project contains a Go language binding for the Open Component Model, which
exactly provides such an extensible implementation frame and a
generic transport functionality based on this mechanism.

Further, the description provides the possibility to add signatures, to
be able to verify the authenticity of the described content even after any
number of transportation steps.

This is achieved by signing a normalized form of a component version description,
which is independent of a concrete serialization format of a [component descriptor](ocm/model.md#component-descriptor)
It includes the digests of the described artifacts, but not the technical access
specifications used to access the artifact in the environment, where the
signature is created.

This way the Open Component model can be used as foundation or some kind of
Lingua Franca for any number of _tools_ dealing with software and software artifacts:
- By using the location-agnostic component, component version and artifact
  identities to denote the entities they are dealing with
- By using the location specific access methods described by a local version
  of the component description to
  - get access to the content of an artifact in defined formats.
  - get the local location of the artifact.
  - verify the authenticity of the artifacts found in the local environment.

Because the identities and the content (but not the location of the content)
are stable after transportation steps, information stored or provided by
different tools accompanied by the notation scheme provided by the Open
Component Model, is exchangeable across different environments and doesn't lose
its validity.

A more detailed specification can be found [here](ocm/README.md).