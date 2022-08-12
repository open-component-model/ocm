# Elements used by the Open Component Model

The Open Component Model provides are formal description of 
delivery artefacts for dedicated semantics that are accessible
in some kind of repository.

This leased to the following major elements that must be specified
the element specification of the Open Component Model

- [Repositories](#repositories)
- [Components](#components)
  - [Component Versions](#component-versions)
    - [Artefacts](#artefacts)
      - [Sources](#sources)
      - [Resources](#resources)
      - [Artefact Access](#artefact-access)
    - [Labels](#labels)
    - [Repository Contexts](#repository-contexts)
    

## Repositories

A *Component Repository* is a dedicated entity that provides technical access
to the other elements of the Open Component Model.

So far, we don't define a dedicated repository API for a dedicated technical
instance of an OCM  repository, because we want to use existing storage
subsystems, without the need of running OCM specific servers.

Therefore, a component repository is typically given by a well-known storage
subsystem  hosting a content structure adhering to an [element mapping specification](interoperability.md) 
for this dedicated kind of storage backend (e.g. OCI).

So, any tool or language binding can map an existing storage technology into an
OCM repository view by implementing the [abstract operations](operations.md)
using this specification for the dedicated storage technology.

If required, an own specification for a native OCM repository (similar to the
[OCI distribution spec](https://github.com/opencontainers/distribution-spec/blob/main/spec.md))
can be added.

## Components

A *Component* is an abstract entity describing a dedicated usage context or
meaning for provided software. It is technically defined by a globally
unique identifier

A component identifier uses the following naming scheme:

<center>

*&lt;DNS domain>* `/` *&lt;name component> {* `/` *&lt;name component> }*

</center>

Hereby the DNS domain plus optional some leading name components MUST 
be owned by the provider of a component.

The component acts as a namespace to host multiple [*Component Versions*](#component-versions),
which finally describe dedicated technical artefact sets, which describe the
software artefacts required to run this tool.

*Example:*

The component with the identity `github.com/gardener/external-dns-management` 
contains software versions of a tool maintaining DNS entries in DNS providers 
based om on Kubernetes resource manifests.

Hereby, the prefix `github.com/gardener` describes a *GitHub* organization owned
by the Gardener team developing the component `external-dns-management`.

### Component Versions

A *Component Version* is a concrete instance of a [Component](#components).
As such it describes a concrete set of [Artefacts](#Artefacts)
adhering to the semantic assigned to the containing Component. It has a unique
identity composed of the component identity and a version name following 
the [semantic versioning](https://semver.org) specification.

So, all versions provided for a component provide software artefacts with the 
same purpose.

A component version is formally described by a [Component Descriptor](#component-descriptor).

#### Component Descriptor

A *Component Descriptor* is used to describe a dedicated component version.
It is a YAML file with the structure defined [here](../formats/compdesc/README.md)

It describes:
- a history of [Repository Contexts](#repository-contexts) describing
  former repository locations of the component version along a transportation
  chain
- a set of [Labels](#labels) to assign arbitrary kinds of information to the
  component version, which is not formally defined by the Open Component Model.
- an optional set of [Sources](#sources), that were used to generate the 
  [Resources](#resources) provided by the component version
- a set of [Resources](#resources) provided with this component version
- an optional set of [References](#references) included component versions
- an optional set of [Signatures](#signatures) provided by some authority 
  to confirm some state or origin of the component version

#### Artefacts

An *Artefact* is a blob containing some data in some technical format.
Every artefact described by the component version has
- an [Identity](#identity) in the context of the component version
- a dedicated globally unique [type](../names/resourcetypes.md) representing
  the kind of content and how it can be used
- a set of [Labels](#labels) to assign arbitrary kinds of information to the
  component version, which is not formally defined by the Open Component Model.
- a formal description of the [Access Method](#artefact-access) which can be used
  to technically access the content of the artefact in form of a blob with a 
  format defined by the artefact type and an optional media type assigned to
  the access method.
- a digest of the artefact that is immutable during transport steps.

##### Identity 

[Resources](#resources), [Sources](#sources), and [References](#references)
have a unique identity in the context of a [Component Version](#component-versions).

All those element types share the same notion of an identity, which is a set
of key/value string pairs. 
This includes at least the value of the `name` attribute of those elements.
Optionally a `version` attribute can be given. If the element name is not
unique in the context of the component version for the actual element type,
the version attribute is added to the element identity. 

Optionally explicit identity attributes can be defined. If given all those
attribute always contribute to the identity of the element.

##### Sources

A *Source* is an [Artefact](#artefacts), which describes the sources that were
used to generate one or more of the [Resources](#resources) described by the
[component descriptor](#component-descriptor).

##### Resources

A *Resource* is an [Artefact](#artefacts), which is a delivery artefact,
intended for deployment into a runtime environment, or describing additional
content relevant for a deployment mechanism, for example installation procedures
or meta-model descriptions controlling orchestration and/or deployment mechanisms.

The Open Component Model makes absolutely no assumptions, how content described
by the model is finally deployed. All this is eft to external tools and tool
specific deployment information formally represented as other artefacts with
an appropriate dedicated own type.

In addition to the common [artefact](#artefacts) information a resource
may optionally describe a reference to the [source](#sources) by specifying
its artefact identity

##### Artefact Access

The technical access to the physical content of an [artefact](#artefacts) described as
part of a [Component Version](#component-versions) is specified by an
*Access Method*. It describes the access path to the content in the 
location the component descriptor has been retrieved from.

Every access method as a formal type and type specific attributes.
The type uniquely specifies the technical procedure how to use the
attributes and the [repository context](#repository-contexts) of
the component descriptor containing the [access method specification](../formats/accessmethods/README.md)
to retrieve the context of the artefact.

#### References

A component version may refer to other [component versions](#component-versions)
by adding a *Reference* to the component version. 

The reference describes only the component version and no location or OCM
repository. It is always evaluated in the context
This means, that the artefact set described by the referenced component version
is added to the local artefact set described by the component version defining
the reference. To keep a unique addressing scheme, like [artefacts](#artefacts),
references have an [identity](#identity). 

Any local or non-local artefact can then be addressed relative to a component
version by a possibly empty sequence of reference identities followed by
the artefact identity in the context of the finally addressed component
version.

The composite is called *Artefact Reference* or *Resource Reference*.
It can be used in artefacts to refer to other artefacts described by the 
same component version containing the artefact.

*Example:*

CompVers: `A:1.0.0`
```
- Resources: 
  - name: DEPLOYER
  - type: mySpecialDeploymentDescription
- 
- References:
  - name: content
    component: B:1.0.0
```

ComVers: `B:1.0.0`
```
- Resources:
  - name IMAGE
    type: ociImage
```

The deployment description contained in CompVers `A:1.0.0` may have
the following content

```
...
deploymentImages:
  - resource:
      name: IMAGE
      referencePath:
      - name: content
```

This description contains a resource reference indicating to
use the resource `IMAGE` in component version `B:1.0.0` when evaluated
in the context of component version `A:1.0.0`.

This way any content related tool can interact with the Open Component Model,
by identifying resources and finally access resources described by the component
model which is agnostic of the evaluation context of the component version.

Depending on the transport history of the component version, always the
correct resource location valid for the actual environment is used.

#### Labels

*Labels* can be used to add additional formal information to a component
model elements which do not have static formal fields in the
[component descriptor](#component-descriptor). Its usage is
free to users of the component model. To assure, that this information
has a globally unique interpretation labels must comply to some naming scheme
and usa a common [structure](../names/labels.md).

# Repository Contexts

A *Repository Context* describes the access to an [OCM Repository](#repositories).

This access is described by a [formal and typed specification](../names/repositorytypes.md)

A [component descriptor](#component-descriptor) may contain information
about the transport history by keeping a list of repository contexts.
It should at least describe the last repository context for a remotely accessible
OCM repository it was transported into.

#### Signatures

