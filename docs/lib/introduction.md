# Introduction to the **ocm-lib**

As the **ocm-lib** is a large and complex library, it can be rather frustrating to get started. This document is
supposed to support you in this process.

## Section 1 - Complexity

As mentioned in the [specification](https://github.com/open-component-model/ocm-spec#storage-technology), an
*OCM Repository* is an interpretation layer on top of existing storage technologies and is not itself a repository
technology.  
Therefore, the ocm-lib has to deal with multiple
[storage technologies](https://github.com/open-component-model/ocm-spec#storage-technology) and their individuel
authentication mechanisms. As it is impossible to forsee all potentially required storage technologies, the ocm-lib's
architecture also has to enable convenient extensibility.  
This leads to a substantial amount of inherent unavoidable complexity!

## Section 2 - Architectural Concepts

Essentially, the ocm-lib revolves around a single architectural concept, called **Type Registries** in the context of
this document.

### Section 2.1 - (Model) Type Registries

To avoid any confusion right from the beginning, a short explanation of the name **Type Registries**. The word _Type_ in
_Type Registry_ does not refer to a _technical type_ within your programming language, e.g. a _go type_ you defined, but
to a _type in your underlying model_.

Suppose you want to write an application for printing messages. Now you know beforehand that you at least have simple
messages that are merely a continuous text and more complex messages that are composed of a title and a text.
[Here](introduction/prototype-based-typeregistry/serializedmessages/) are examples how their respective serializations
may look
like. They both have a _type_ - `type: simplemessage` and `type: complexmessage`. These typenames are unique and
describe their **model types**. Now, _Type Registries_ are essentially maps that map a _model type_ (key) to a construct
in your programming language that can properly deal with that _model type_ (value). This might actually be dedicated go
type for each model type (as in the prototype-based type registry), but this might just as well be objects of the same
type with a dedicated configuration (as in the factory-based type registries).

#### Section 2.1.1 - Prototype-Based Type Registry

_This kind of Type Registry only works if there is a **dedicated go type** for each **model type**, and if it is
sufficient to directly decode
the serialized data into an empty object of that go type without requiring further processing._

The [prototype-based-typeregistry](introduction/prototype-based-typeregistry) directory contains a quite minimal example
implementation of a single _Prototype-Based Type Registry_ that will be used as reference to explain the concept
(instead of reading this lengthy explanation, you can also just check out the well-documented code). The example creates
a _Type Registry_ for types implementing the
following [Message](introduction/prototype-based-typeregistry/registry/interface.go)
interface.

    type Message interface {
        Print()
    }

In _Prototype-Based-Type-Registries_, the Registry is essentially a map that stores the model type name (key) and a
prototype object of a dedicated go type implementing the corresponding functionality (value). This prototype object is
used to create new objects of that type  (leveraging the reflection library).

    type MessageRegistry map[string]Message

Typically, the [package defining the type registry](introduction/prototype-based-typeregistry/registry/registry.go) also
declares and
initializes a _DefaultRegistry_, a global variable that can be used by other packages to register their implementation
of the respective type (here, Message).

    var DefaultMessageRegistry = MessageRegistry{}

To register themselves at the _DefaultRegistry_, these other packages implement their
[init-function](https://go.dev/doc/effective_go#init) respectively (thereby, the type name should obviously be unique).

    const TYPE = "simplemessage"

    func init() {
        registry.DefaultMessageRegistry[TYPE] = &Message{}
    }
    
    type Message struct {
        Type string `json:"type"`
        Text string `json:"text"`
    }
    
    func (m *Message) Print() {
        fmt.Println(m.Text)
    }

Keep in mind that although the
[_simplemessage_](introduction/prototype-based-typeregistry/types/simplemessage/message.go) and
[_complexmessage_](introduction/prototype-based-typeregistry/types/complexmessage/message.go) packages might not be
directly used in other packages, they have to be imported for their side effects (in other words, to execute their
init-functions). You may use empty imports to achieve this.

    import (
        _ "example/typeregistry/types/complexmessage"
        _ "example/typeregistry/types/simplemessage"
    )

Now, the prepared _DefaultRegistry_ can be used to dynamically unmarshal serializations of the types into their
respective go types.

The [registry](introduction/prototype-based-typeregistry/registry/registry.go) has a function implementing the dynamic
unmarshaling logic. To understand this, it's best to check out the well-documented method within the example.

#### Section 2.1.2 - Factory-Based Type Registry

In _Factory-Based Type Registries_, the Registry is essentially a map that stores the model type name (key) and a
factory object (thus, an object capable of producing Message objects). Thereby, it is possible to overcome the
limitations of the Prototype-Based Type Registries (the necessity of a dedicated go type per model type and not being
able to edit the object after unmarshaling).

Based on the knowledge about Prototype-Based Type Registry, it should be possible to understand the well-documented
[Factory-Based Type Registry](introduction/factory-based-typeregistry) example.

#### Note

The ocm-lib uses has several kinds of type registries (thus, corresponding to our example, not only a Message Registry
but also for other types). There, it has a generic implementation of a registry which can be reused.  
Furthermore, its implementation of the registry has to be thread safe and provides some convenience functionility.
Therefore, it may be slightly more difficult to understand than our examples.

#### Section 2.1.3 - Purpose

The concept and implementation of _Type Registries_ may initially seem unnecessary complex and cumbersome. But the power
of this concept becomes really apparent in the context of the [ocm](https://github.com/open-component-model/ocm-spec),
once one realizes that the whole standard revolves around
[Component Descriptors](https://ocm.software/docs/component-descriptors/version-3/) that are essentially bundled lists
of serialized typed (and versioned) objects.

Furthermore, the concept is great for extensibility. To create a new type, you can theoretically just create a new
self-contained package implementing this type and register it at the corresponding _DefaultRegistry_ and you are good to
go (so the code perfectly abides to the _open-closed principle_).

### Section 2.2 - Contexts

The ocm-lib defines multiple _Context_ types. The documentation of the library states that objects of these _Context_
types bundle all settings and extension point implementations for this area.

**Extension Point Implementations:** On a conceptual level, each package providing a type that implements an interface
that has a corresponding _Type Registry_ is considered an extension point implementation (in the previous
examples, the _simplemessage package_ and the _complexmessage package_ are extension points as they
implement the _Message_ interface and there is a _Message Type Registry_).  
**Settings:** On a conceptual level, settings are values that may be set for specific objects or passed to specific
functions as an attribute to influence the behavior, e.g. a repository object might have fields username and password
that can be set and will be used by the respective access method on that object to access the repository **or** the
username and password may be passed directly to the respective function (in the [context](introduction/context) example,
the print settings are settings as they are used within the print methods of functions).

So, on a conceptual level, _Context_ objects are containers for a number of _Type Registries_ and _Settings_ (as
shown in the [context](introduction/context) example). They provide several methods that rely on that information (for
example, the dynamic unmarshaling, as this depends on the types (or rather extension points) registered in registry).
Furthermore, they are passed into other objects, so that they can access the respective settings (like the Message
objects access the print settings within the example).

## Section 3 - Mapping the concepts to the ocm-lib

The following explanations shall help mapping the just explained concepts to their usage in the ocm-lib and to get a
general idea of the directory structure.

### Section 3.1 - Overview

The picture below shows the basic structure of the ocm-lib's [ocm context package](../../pkg/contexts/ocm). As all
context packages follow the same structure, the overview should be sufficient to navigate the other context packages.

The elements within this overview (_Type Registries_ and _Contexts_) are explained below.

![image](introduction/ocm-lib-structure.png)

### Section 3.2 - Type Registries

Here, the different _Type Registries_ shown in the overview are explained.

### Section 3.2.1 - Repository Types Scheme

A `RepositoryTypesScheme` is an object that maps a number of _repository types (= model type)_ to corresponding 
_factories_ providing a repository object of a specific type or with a specific configuration for dealing with this 
_repository type (= model type)_.

There exist multiple _Repository Spec Types_ Registries within the ocm-lib. There is a Repository Spec Type Registry
for [**OCM** _repository types (= model types)_](../../pkg/contexts/ocm/repositories) and a Repository Spec Types
Registry for [**OCI** _repository types (= model types_](../../pkg/contexts/oci/repositories).

**OCI Repository Types:**  
Having multiple _OCI Repository Types_ may initially sound confusing. Therefore, a short explanation on what
_OCI Repositories_ and _OCI Repository Types_ even are in the context of the ocm-lib. From the perspective of the 
ocm-lib, any _go type_ that implements the following [interface](../../pkg/contexts/oci/internal/repository.go) can be 
considered an _OCI Repository_:  

```
type Repository interface {
   GetSpecification() RepositorySpec
   NamespaceLister() NamespaceLister

   ExistsArtifact(name string, ref string) (bool, error)
   LookupArtifact(name string, ref string) (ArtifactAccess, error)
   LookupNamespace(name string) (NamespaceAccess, error)
   Close() error
}
```

Within the ocm-lib, there are multiple implementations of this interface. Most prominently, an implementation for actual
[_OCI Registries_](../../pkg/contexts/oci/repositories/ocireg). If you are familiar with _OCI Registries_, the
terminology here may be counterintuitive. In the ocm-lib, the term _OCI Repository_ corresponds to what is commonly 
known as an _OCI Registry_, and a _Namespace_ corresponds to what is commonly known as an _OCI Repository_.  
Another implementation is the so called _[Common Transport Format](../../pkg/contexts/oci/repositories/ctf)_. The 
_[Common Transport Format specification](https://github.com/open-component-model/ocm-spec/blob/main/doc/appendix/A/CTF/README.md)_ 
was developed in the context of the Open Component Model and describes a file system structure that can be used for the 
representation of content of an OCI Registry.  
As one might notice within the [package](../../pkg/contexts/oci/repositories/), besides `OCIRegistry` and `Common
Transport Format`, there are further _OCI Repository Types_, namely, `DockerDaemon` and `ArtifactSet`. These are not 
"equal" implementations compared to `OCIRegistry` and `CommonTransportFormat`. The _docker daemon_ is limited to
handling _OCI Images_ and therefore cannot deal with _OCI Artifacts_. An _artifact set_ can only hold versions of the same
artifact (thus, it corresponds to a Namespace or rather OCI Repository in that regard), but implements the _OCI
Repository_ interface. 

**OCM Repository Types:**
So, corresponding to OCI Repositories and OCI Repository Types, any _go type_ that implements the following 
[interface](../../pkg/contexts/ocm/internal/repository.go) can be considered an _OCM Repository_:

```
type Repository interface {
	GetContext() Context

	GetSpecification() RepositorySpec
	ComponentLister() ComponentLister

	ExistsComponentVersion(name string, version string) (bool, error)
	LookupComponentVersion(name string, version string) (ComponentVersionAccess, error)
	LookupComponent(name string) (ComponentAccess, error)

	Close() error
}
```

Again, within the ocm-lib, there are multiple implementations of this interface. But most importantly, the package
[`genericocireg`](../../pkg/contexts/ocm/repositories/genericocireg) provides an implementation of the _OCM Repository 
interface_ that is based on the _OCI Repository interface_. In other words, this package provides a mapping of the
_OCM Repository_ functionility to the _OCI Repository_ functionality. The 
[`ocireg`](../../pkg/contexts/ocm/repositories/ocireg) package and the [`ctf`](../../pkg/contexts/ocm/repositories/ctf) 
merely provide the functionality to convert, or rather wrap, their _OCI Repository Spec_ to an _OCM Repository Spec_
(remember, spec types were pretty much serializable representations and factories for the actual type, as explained
[here](introduction/context/registry/registry.go)).  
For each _OCI Repository Type_, there also exists a corresponding _OCM Repository Type_.
Thus, as there is an **oci** _repository type_ `OCIRegistry`, there is also an **ocm** _repository type_ `OCIRegistry`, 
as there is an **oci** _repository type_ `CommonTransportFormat`, there is also an **ocm** _repository type_ 
`CommonTransportFormat`. 
Within the ocm-lib this is implemented through an **addition to the factory-based type registry** concept. In case the
does not find an entry for a certain type (within the Decode method), this 
[implementation (defaultScheme)](../../pkg/runtime/scheme.go) of a factory-based type registry has a _fallback factory_ 
(called `defaultdecoder`). For the OCM Repository Type Registry, this _fallback factory_ is usually 
[initialized](../../pkg/contexts/ocm/builder.go) with a Type Registry itself, the 
[`DefaultDelegationRegistry`](../../pkg/contexts/ocm/repositories/genericocireg/type.go) which is based on the 
[`delegationRegistry`](../../pkg/contexts/ocm/internal/delegation.go). This Type Registry's iterates through all its 
factories and checks whether they knew that type. Currently, the only factory known by this _fallback Type Registry_ is
the one registered in the [`genericocireg`](../../pkg/contexts/ocm/repositories/genericocireg/type.go) package which
uses the **oci** _repository types registry_ to perform the unmarshaling (thus, if I registered the type in the 
oci repository types registry, it is automatically available for the ocm repository registry).

### Section 3.2.2 - Access Methods

A AccessMethods object stores a number of types implementing a AccessMethod (thus, AccessMethod is the "certain
functionality").
In this case, one type is ociArtifact (storing the artifact in an oci Registry), thus, a dedicated technical procedure
of how to access the artifact blob described by a Component Descriptor. Further types are localBlob (storing the
artifact blob along with the component descriptor) or github (storing the artifact blob in a GitHub registry).
To make it even more clear, an Access Method is anything that allows to access artifact blobs described by a Component
Descriptor through the following interface (thus, describes the following functionality):

    type AccessMethod interface {
        Get() ([]byte, error)
        Reader() (io.ReadCloser, error)
        
        GetKind() string
        AccessSpec() AccessSpec
        MimeType() string
        
        Close() error
    }

CAREFUL - There are two conceptually relevant details here!

1) There are global and local AccessMethod types. Global types (such as ociArtifact or github) implement the procedure
   to access the artifact blobs themselves. Local types, on the contrary, delegate the implementation of the procedure
   to the respective component repository (as each type of component repository may handle the storage of artifact blobs
   alongside the component descriptor differently).
2) AccessMethods return the artifact blob as is. Thus, if a helm chart is stored as an .tar, the AccessMethod will
   return the bytes of the .tar. If, on the other hand, the helm chart is stored as an tar.gz , the AccessMethod will
   return the bytes of the .tar.gz. Therefore, for further processing, the MimeType() can be read (inspecting the
   MimeType and conversion to a target format can be done with downloaders - a convenience feature implemented by the
   ocm library kind of as an add-on).

### Section 3.2.3 - Repository Spec Handlers

A RepositorySpecHandlers object stores a number of types implementing a RepositorySpecHandler (thus,
RepositorySpecHandler is the "certain functionality").

It is rather cumbersome to enter JSON objects such as a RepositorySpec through the command line. Furthermore, there may
be multiple types of RepositorySpecs for certain type of repository (e.g. legacy types).
To deal with this issue, string representation of a repository spec (e.g.
eu.gcr.io/gardener-project/landscaper/examples/charts/hello-world:1.0.0)  may be used and parsed into a
UniformRepositorySpec. Based on the type attribute of the UniformRepositorySpec, a specific RepositorySpecHandler is
chosen to map the UniformRepositorySpec to a specific RepositorySpec.
To make it even more clear, a RepositorySpecHandler is anything that allows to map a UniformRepositorySpec to a
RepositorySpec through the following interface:

    type RepositorySpecHandler interface {
        MapReference(ctx Context, u *UniformRepositorySpec) (RepositorySpec, error)
    }

### Section 3.2.4 - Blob Handlers

A BlobHandlers object stores a number of types implementing a BlobHandler (thus, BlobHandler is the "certain
functionality").

When a Component Version, or rather, a Component Descriptor, is imported into an OCM Repository, a decision has to be
made about how to deal with the artifacts, or rather, the artifact blobs (e.g. does it stay a local artifact stored
alongside the Component Descriptor or is it uploaded into an OCI Registry as an independent OCI Artifact?).
These decisions can be made within BlobHandlers. BlobHandlers therefore implement the following interface:

    type BlobHandler interface {
        StoreBlob(blob BlobAccess, artType, hint string, global AccessSpec, ctx StorageContext)   (AccessSpec, error)
    }

The function has to return an AccessSpec that can then be inserted into the respective new Component Descriptor.

### Section 3.2.5 - Blob Digesters

A BlobDigesters object stores a number of types implementing a BlobDigester (thus, BlobDigester is the "certain
functionality").

In theory, the digest could be calculated over the byte sequence of the blob. But e.g. for archives, the digest may
change after through unpacking and repacking processes (dates may be added or changed, the order of specific elements
may change). As a consequence, the same artifact could have different digests which would be a problem regarding
signing. Therefore, there exist digesters for specific mime types e.g. an OCI artifact digester, that calculates the
digest over the byte sequence of the OCI artifact representation which always has to be the same per specification.
Generally, a BlobDigester is anything that allows to calculate the digest for blobs with specific mime types through the
following interface:

    type BlobDigester interface {
        GetType() DigesterType
        DetermineDigest(resType string, meth AccessMethod, preferred signing.Hasher)   (*DigestDescriptor, error)
    }

## Section 3.3 - Dictionaries

As explained previously, _Type Registries_ are essentially maps that map a _model type_ (key) to a construct in your
programming language that can properly deal with that _model type_ (value). _Dictionaries_, on the contrary, do not
deal with _model types_ but rather arbitrary _IDs_ (e.g. ConsumerIdentities) to a corresponding construct in your
respective programming language.

### Section 3.3.1 - Identity Matchers

An IdentityMatchers object stores a number of types (in this case functions) implementing an IdentityMatcher (thus,
IdentityMatcher is the "certain functionality").

Essentially, an IdentityMatcher is a function implementing the following interface:

    type IdentityMatcher func(pattern, cur, id ConsumerIdentity) bool

A ConsumerIdentity is essentially just a `map[string]interface{}`. There are 2 standard matchers, a PartialMatcher and a
CompleteMatcher. The CompleteMatcher returns true only if the pattern and the id are completely equal. The
PartialMatcher returns true, if the id matches the pattern partially AND better, thus in more attributes, than curr.
This allows to iterate over several available Consumers (these consist of an ConsumerIdentity and Credentials) and find
the best match, and thereby, hopefully, the correct Credentials.

### Section 3.3.2 - Consumer Providers

A ConsumerProviders object stores a number of objects called ConsumerProviders which are essentially stores for
Consumers (Thus, contrary to a lot of other registries that store types implementing a specific interface,
ConsumerProviders stores actual instances of types, in other words objects, implementing a specific interface).

A ConsumerProvider has to implement the following interface:

    type ConsumerProvider interface {
        Unregister(id ProviderIdentity)
        Get(id ConsumerIdentity) (CredentialsSource, bool)
        Match(id ConsumerIdentity, cur ConsumerIdentity, matcher IdentityMatcher
            (CredentialsSource, ConsumerIdentity)
    }

A Consumer is essentially the combination of an Identity (the ConsumerIdentity) and a CredentialSource (which is
essentially something that can provide some sort of map containing the credentials).

The advantage of having ConsumerProviders instead of adding all credentials to a central store within the
CredentialContext is that each CredentialRepository can simply bring its own specific ConsumerIdentity matcher.
Furthermore, to in order to update the propagated credentials, the ConsumerProvider may simply be exchanged. If all
Consumers would be stored centrally in the CredentialContext, the Consumers propagated by a specific repository would
have to be identified and removed.

## Section 3.4 - Contexts

Here, the different registries shown in the overview are explained.

A context object is the entry point for using dedicated functional areas. It bundles all settings and extensions point
implementations for this area.

This becomes more tangible if one looks at a concrete instance of such a functional area, e.g. ocm. The OCM Context
contains several registries which determine e.g. which repository types or access methods are known and thus, can be
used. So this may be seen as settings.
Furthermore, by implementing e.g. a new type of repository and registering at the repository types registry of the
current context, one may extend the functionality of the library.

A Context may be seen as entry point as it provides access to a set of basic root element of its functional area. This
is typically a Repository object, which provides access to elements hosted by this repository. An OCMRepository provides
access to Components and ComponentVersions. An OCIRepository provides access primarily to so called Namespaces. A
Namespace provides access to versions of an OCI Artifact.
As you might have noticed in this description, the term "Namespaces" is unusual here and is an oddity of this ocm-lib.
What is referred to as Namespace or rather OCI Namespace in this library is usually referred to as OCI Repository and
correspondingly, what is referred to as OCI Repository in this library is usually referred to as OCI Registry. To be
even more concrete, the library has following mapping of terms:

    Name in OCM Library             Name used everywhere else
    OCI Repository          ->      OCI Registry
    OCI Namespace           ->      OCI Repository

### Section 3.4.1 - Data Context

The DataContext type is kind of base context of all Context types. Therefore, every Context type (except the Data
Context itself) contains a DataContext object.  
Essentially, a DataContext object contains the type information of the Context object it is the base for (e.g.
ocm.context.ocm.software in case of an OCM Context) and a specific attribute store (and internally also a pointer to the
Context object it is the base for).
(The data context is an AttributeContext itself and therefore, the attribute store of the DataContext works the same as
explained in AttributesContext.)

### Section 3.4.2 - Attributes Context

Every Context object - besides the DataContext, which is itself an AttributesContext - contains an AttributesContext
object. An AttributesContext also contains a DataContext object. Primarily, an AttributesContext provides an attribute
store which can store arbitrary attributes (on a low level, it this is implemented through a map[string]interface{}).
These arbitrary attributes are typed and the respective types have to enable marshaling and unmarshaling, so that the
attributes can also be used on the cli. In other words, an attribute type has to implement the following interface:

    type AttributeType interface {
        Name() string
        Decode(data []byte, unmarshaler runtime.Unmarshaler) (interface{}, error)
        Encode(v interface{}, marshaller runtime.Marshaler) ([]byte, error)
        Description() string
    }

So, before an attribute is stored in an AttributesContexts attribute store, it is checked, whether its type is
registered as AttributeType at a global registry called DefaultAttributeScheme. As every type of Context object contains
a AttributeContext object, functions on Context types can try to access attributes in the AttributesContext (e.g. The
StoreBlob method of an BlobHandler accesses an ociuploadattr on the DataContext. This attribute can specify an OCI
registry where all the artifacts shall be uploaded to and thereby may overwrite the default behaviour of uploading the
artifacts to the same OCI registry as the Component Descriptor.)
Attributes stored in the AttributesContext of a context overwrite corresponding attributes stored in the DataContext of
the same context.

### Section 3.4.3 - Config Context

A ConfigContext object configures other configurable objects (e.g. Credential Context) using typed configuration
objects (thus, the object has a unique type attribute specifying its own type, e.g. credentials.config.ocm.software).

Conceptual Background - Config Objects (this is pretty much how it works, but for reasons that will become clear later
on, the ApplyTo function interface has an additional parameter compared to how it's introduced here):
Each Config object type has to have an ApplyTo(target interface{}) function that knows what type of configurable object(
s) it can configure (this "knowledge" is typically implemented through an if-statement based on a type-assertion on the
target). If the function receives a suitable configurable object as target (thus, the type-assertion returns true), it
calls corresponding functions on the target to apply its respective configuration (e.g. , a Credential Config object
calls the credentialContext.SetCredentialsForConsumer(...) function passing in the respective arguments).

To conclude this, technically, in order to configure specific objets such as the credential context - let's call it
credctx - one could use a ConfigObject of the corresponding type, here Credential Config - let's call it credcfg - and
call credcfg.ApplyTo(credctx).

[Typically, each directory that contains configurable objects also has a corresponding config package with a type.go file, defining the Config type and registering it at the default ConfigTypeScheme.]

Conceptual Background - ConfigContext (this is, where it might get confusing, but bear with me):
The ConfigContext object is a configurable object itself. The Config object type to configure the ConfigContext is
generic.config.ocm.software. This Config object is pretty much only a container for Config objects (thus, it simply
contains a list of arbitrarily typed Config objects). Calling the ApplyTo(target interface{}) function of such an object
with a ConfigContext as target adds all the Config objects contained in this container Config object to a ConfigStore
within the ConfigContext.

ConfigContext (how the library intends to work with Config objects):
The Config objects are usually not applied directly to one or multiple configurable objects (thus, calling the Config
objects ApplyTo(target interface{}) passing in each configuration object as target).
Instead, a ConfigContext provides an (c ConfigContext) ApplyConfig(spec Config, description string) function.
Essentially, this function does two things:

1) Call spec.ApplyTo(target interface{}), passing in the ConfigContext its called upon in as target
2) Adding the spec (thus, the Config object) to the ConfigStore of the ConfigContext

Here's an example to better grasp this behaviour:
Let cfgctx be an object of type ConfigContext and credcfg be an object of type CredentialConfig object.
Assume the following function call:
cfgctx.ApplyConfig(credcfg, "credential")

1) Now, as previously described, the function first calls credcfg.ApplyTo(cfgctx). As the credcfg object's ApplyTo
   function knows that it can only configure objects of type CredentialContext and not of type ConfigContext, this
   function call does not have any effect!
2) After the ApplyTo function returns, regardless of the type of the Config object, it is added to the ConfigStore of
   the cfgctx.

Now, let cfgctx, again, be an object of type ConfigContext and, instead of credcfg, we now have a cfgcfg object of type
ConfigConfig (generic.config.ocm.software, the Config object type to configure ConfigContext objects).
Assume the following function call:
cfgctx.ApplyConfig(cfgcfg, "config")

1) Now again, the function first calls cfgcfg.ApplyTo(cfgctx). As the cfgcfg object's ApplyTo function knows that it can
   configure objects of type ConfigConfig, this function call proceeds and adds all all configuration object within the
   cfgcfg object (remember, it's a container for other configuration objects) to the ConfigStore of the cfgctx.
2) After the ApplyTo function returns, the cfgcfg object itself (thus, the container) is added to the cfgctx.
   (if the cfgcfg object contained 3 different Credential Config objects, the ConfigStore would now contain the 3
   Credential Config objects and also the cfgcfg object, thus, 4 objects in total)

"Cool, now I've got an ConfigContext object containing all my configuration objects, but what's the benefit? How are
they applied to the actual objects they shall configure?"

Updater (this is an internal detail, a user of the library does not have to deal with the updater, but it's important if
you want to understand, how the configuration objects within the ConfigStore of a ConfigContext are applied to the
actual configurable objects!):
Configurable Objects (e.g. OCM Context, OCI Context, Credential Context) do not directly contain a ConfigContext object
or rather a pointer to a ConfigContext object. Instead, they contain an Updater object which maintains a pointer to a
ConfigContext. Essentially, the Updater consists of the triple (ConfigContext, Target, LastGeneration).
The Target is a configurable object, typically the configurable object that contains the respective Updater.
The LastGeneration is an integer. The ConfigContext also maintains a corresponding Generation integer, which is
increased by 1 each time a new Config object is added to the ConfigStore.
Every method of a configurable object that uses a potentially configurable part of the configurable object calls its
Updaters Update() function before performing any other processing. This Update() function compares the ConfigContext's
Generation with the LastGeneration and calls the ApplyTo method of each Config object added since the LastGeneration,
passing in the Updaters target object as target.
This mechanism allows the transparent configuration of objects through a uniform interface.

**FAQ:**  
Why does the ConfigContext store the Container Config object (generic.config.ocm.software) in its ConfigStore?  
-> The ConfigContext does not know anything about this Container Config object type. There could be further types of
Config objects that configure the ConfigContext (and eventually even other types of configurable objects too). These
would have to be stored in the ConfigStore. Therefore, not storing the Container Config objects would require special
treatment, which is not necessary.

Why do the ConfigObjects apply themselves to the configurable objects and not the other way around? In other words, why
don't the configurable objects provide an apply method?  
-> As the configurable objects do not have to know anything about their configuration objects, it is easy to create an
additional Config object (This type would only have to register itself to the respective ConfigTypes Registry and it
would be good to go!)

### Section 3.4.4 - Credential Context

A CredentialContext object serves two purposes:

1) it stores Consumers (indirectly in form of Consumer Providers)
2) it provides access to Credential Repositories (just like the OCM Context provides access to OCM Repositories or the
   OCI Context provides access to OCI Repositories)

**Credential Management:**  
There are several kinds of credential consumers (e.g. GitHub, HelmChartRepositories, OCIRegistries) with potentially
completely different kinds of credentials (e.g. {key:<GitHubPersonalAccessToken>} in case of GitHub or
{username:<basic auth name>, password:<basic auth password>, certificate:<TLS client certificate>,
privateKey:<TLS private key>} in case of HelmChartRepositories).

To account for these different kinds of consumers and credentials, the credential management is based on generic
consumer identities and generic credential property sets.
The generic consumer identities are composed of a consumer type and a set of properties further describing the context
the credentials are required for. The consumer type specifies a dedicated type of consumer (e.g. GitHub,
HelmChartRepositories, OCIRegistries) and thereby also implicitly determines a standard identity matcher (These standard
identity matchers are typically registered at a globally variable called StandardIdentityMatchers during the
initialisation. When creating a DefaultContext, this variable is assigned to the IdentityMatchers Registry of the
CredentialContext.). The set of properties specifying the usage context usually depends on the consumer type. consumer
identities with the consumer type Github typically specify hostname, port and potentially a pathprefix.
The generic credential property sets generally also depend on the consumer type. But even for a concrete consumer type
they may vary (e.g. OCIRegistry allows authentication with username and password, but also with an identity token).
Therefore, after retrieving the Credentials through the ConsumerIdentity matching logic, the consumer implementation may
have to check which kind of credentials it has (e.g. the OCIRegistry implementation tries to retrieve a property with
the name identityToken and if this returns nothing, it tries to retrieve properties with the name username and
password).

**Additional Notes:**  
In theory, a Credential Repository could itself require credentials to log into (e.g. vault). That's what the Credential
Chain is intended for (currently not used)

### Section 3.4.5 - OCI Context

The OCI Context provides access to OCI Repositories (or rather OCI Registries).
The primary aspect to consider here, is that OCI Repository is just an interface, thus, an abstraction. In other words,
any type that provides access to OCI Artifacts through the OCI Repository interface is considered an OCI Repository by
this library.

Current Implementations (also called Storage Mappings) are:  
**ocireg Package**  
The ocireg package within pkg/contexts/oci/repositories implements the OCI Repository interface using OCI Registries

**docker Package**  
The docker package within pkg/contexts/oci/repositories implements the OCI Repository interface using docker daemons

**ctf Package**  
The ctf package within pkg/contexts/oci/repositories implements the OCI Repository interface using file systems.

**CAREFUL:** To prevent confusions, artifactset is not an equal implementations to those. An artifactset can only hold
versions of the same artifact. Thus, it corresponds to a Namespace (or rather OCI Repository), but implements the OCI
Repository interface. The artifactset type is the type in which local artifacts are stored alongside their Component
Descriptor.

### Section 3.4.6 - OCM Context

The OCM Context provides access to OCM Repositories.
The primary aspect to consider here, is that the OCM Repository implementations are based on the OCI Repository
interface. Consequently, it is possible to dynamically choose a different storage backend for the OCM Repository.

The implementation of the mapping from OCM Repository (and respective nested functionalities) to OCI Repository is done
by the genericocireg package in pkg/contexts/ocm/repositories.
Therefore, there are also corresponding packages within pkg/contexts/ocm/repositories for each OCI Repository
implementation that can be used as storage backend for an OCM Repository - to be more precise, these currently are the
ocireg implementation and the ctf implementation (docker, or rather the docker daemon, cannot be used as storage backend
due to the technical limitation that docker can only deal with OCI images and not OCI artifacts).
These packages are rather trivial as each of them pretty much only provide a function to create a genericocireg
RepositorySpec (which implements the ocm RepositorySpec interface) from its own RepositorySpec (which only implements
the oci RepositorySpec interface).

**Additional Notes:**  
In the future, it is quite possible that there will be additional storage backend abstractions besides OCI. To be more
concrete, to be able to store OCM Repositories in a S3 bucket, it is currently considered whether an complete additional
storage backend abstraction, potentially called BlobStore, should be added - parallel to OCI.
Alternatively, it would also be possible to provide a concrete implementation of a OCM Repository for S3 buckets (
without introducing a fully-fledged additional storage abstraction). 






