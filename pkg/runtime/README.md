# Serialization and deserialization of formally typed objects

This package provides support for the de-/serialization of objects into/from a JSON or YAML representation. 

Objects conforming to this model are called *typed objects*. They have a formal type, which
determines the deserialization method. To be able to infer the type from the serialization format, it always contains a formal field `type`. So, the external format of a such an object would be (in JSON):

```
{ "type": "my-special-object-type" }
```

## Core model types

### Simple types objects

*interface `TypedObject`* is the common interface for all kinds of typed objects. It provides access to the name of the type of the object.

*interface `TypedObjectType`* is the common interface all type objects have to implement.
It provides information about the type name and a decode method to deserialize an external
JSON/YAML representation of the object. Its task is to handle the deserialization of
instances/objects based of the type name taken from the deserialization format. 

### Versioned Types

A versioned type is described by a type name, which differentiates between 
a common kind and a format version. Here, there is an *internal* program facing
representian given by a Go struct, which can be serialized into different format
versions described by a version string (for example `v1`). Version name and kind are
separated by a slash (`/`). A type name without a version implies the the version `v1`.

*interface `VersionedTypedObject`* is the common interface for all kinds of typed objects, which provides versioned type. 

*interface `VersionedTypedObjectType`* is the common interface all type objects for VersionedTypedObjects have to implement.

### Schemes

*interface `Scheme`* is a factory for types objects, which hosts descrialization methods (interface `TypedObjectDecoder`) for dedicated types of typed objects. 
type object implement such an interface.


*type `ObjectTypedObject`* is the minimal implementation of a typed objects implementing the
`TypedObject` interface.


## Used Data types

*type `ObjectType`* is the serializable implementation of a type accessor based on a 
type name field.