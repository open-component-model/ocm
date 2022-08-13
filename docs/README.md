# Open Component Model (OCM)

The task of the open component model is to describe a machine read-able formal 
software-bill-of-deliveries (SBOD) with the focus on the delivery
artefacts.

It is a completely technology-agnostic model to describe artefacts and
the technical access to their content. Technology-agnostic means:

- it can describe any kind of artefact regardless of its technology
- the artefacts can be stored in any kind of storage backend or repository
- the model information can be stored in any kind of storage backend.

The only constraint is, that there must be
- an implementation for accessing artefacts in the desired repository technology 
  and map them to a blob format
- and a specification for a [mapping scheme](ocm/interoperability.md) describing how
  to map the elements
  of the component model to the supported elements of the backend technology
- and an [implementation](ocm/operations.md) of all the mapping schemes for the
  storage scenarios used in a dedicated environment.

The model uses a globally unique naming scheme for software [components](ocm/model.md#components).
Components are versioned, and every component version describes
a set of formally typed delivery artefacts (like OCI images) with unique 
identities in the context of the component version.

Those artefact definitions may carry an additional arbitrary attribution, and 
they provide a formal specification of the access method, which can be used
to access the technical content of the artefact in the actual evaluation
context of a component version.

The description model allows to transport content from one repository 
landscape (hosting the real technical artefacts) into other, especially private,
environments without losing the validity of the access information. In any
environment the actual description of the component version carries valid
environment-specific information for the artefact location.

A transport tool, can use the bill of material to determine the set of
artefacts that have to be transferred and use the access information to access
the technical content of the artefacts in the source environment. They will then
be copied into a repository landscape on the target side. In the target
environment, the model description will be stored, also, after it has been
adapted to reflect the new location of the described artefacts.

Using provided implementation for the used access types, this can be done 
in completely generic manner.

Further, the description provides the possibility to add signatures, to
be able to verify the authenticity of the described content even after any
number of transportation steps.

This is achieved by signing a normalized form of a component version description,
including the digests of the described artefacts, but not the technical access
specifications used to access the artefact in the environment, where the
signature is  created.

This way the Open Component model can be used as foundation or some kind of
Lingua Franca  for any number tools dealing with software and software artefacts:
- By using the location-agnostic component, component version and artefact
  identities to denote the entities they are dealing with
- By using the location specific access methods described by a local version
  of the component description to
  - get access to the content of an artefact in defined formats
  - get the local location of the artefact
  - verify the authenticity of the artefacts found in the local environment.

Because the identities and the content (but not the location of the content)
are stable after transportation steps, information stored or provided by
different tools accompanied by the notation scheme provided by the Open
Component Model, is exchangeable across different environments and don't lose
its validity.

A more detailed specification can be found [here](ocm/README.md).