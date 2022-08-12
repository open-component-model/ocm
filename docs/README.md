# Open Component Model (OCM)

The task of the open component model is to describe a machine read-able formal 
software-bill-of-deliveries (SBOD) with the focus on the delivery
artefacts.

It uses a globally unique naming scheme for software [components](ocm/model.md#components).
Components are versioned, and the every component version describes
a set of formally typed delivery artefacts (like OCI images) with unique 
identities in the context of the component version.

Those artefact definitions may carry an additional arbitrary attribution, and 
they provide a formal specification of the access method, which can be used
to access the technical content of the artefact in the actual evaluation
context.

The description model allows to transport content from one repository 
landscape (hosting the real technical artefacts) into other, especially private,
environments without losing the validity of the access information.

A transport tool, can use the bill of material to determine the set of
artefacts that have to be transferred and use the access information to access
the technical content of the artefacts. In the target environment, the 
model description will be stored, again, after it has been adapted to reflect
the new location of the described artefacts.

The description provides the possibility to add signatures, also, to
be able to verify the authenticity of the described content even after any
number of transportation steps.

This is achieved by signing a normalized form of a component version description,
including the digests of the described artefacts, but not the technical access
method used to access the artefact in the environment, where the signature is
created.

This way the Open Component model can be used as foundation or some kind of
Lingua Franca  for any number tools dealing with software and software artefacts:
- By using the location-agnostic component, component version and artefact
  identities to denote the entities they are dealing with
- By using the location specific access methods described by a local version
  of the component description to get access to
  - get access to the content
  - get the local location of the artefact
  - verify the authenticity of the artefacts found in the local environment.

Because the identities and the content (but not the location of the content)
are stable after transportation steps, information stored or provided by
different tools accompanied by the notation scheme provided by the Open
Component Model, is exchangeable across different environments and don't lose
its validity.

A more detailed specification can be found [here](ocm/README.md).