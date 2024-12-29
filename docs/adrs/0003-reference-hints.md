# ADR-0003 - Rework of Reference Hints

- Status: proposed
- Date: 2024-12-27
- Authors: @mandelsoft
- Deciders: @fabianburth @jakobmoellerdev
- Reference Implementation: #1212

### Meaning of Reference Hints

During the transport of software artifacts referenced from external artifact repositories like
OCI registries, they might be stored as  blobs along with the component version (access method
`localBlob`). If those  component versions are transported again into a repository landscape they
might be uploaded again to external repositories.

To provide useful identities for storing those artifacts hints in external repositories, again,
the original identity of the external artifact must be stored along with the local blob.

### Current Solution

The access method originally used to reference the external artifact provides a reference hint,
which can later be used by the blob uploaders to reconstruct a useful identity.
Therefore, the `localBlob` access method is able to keep track of this hint value.
The hint is just a string, which needs to be interpreted by the uploader.

### Problems with Current Solution

The assumption behind the current solution is that the uploader will always upload the
artifact into a similar repository, again. Therefore, there would be a one-to-one relation
between access method and uploader.

Unfortunately this is not true in all cases:

- There are access methods now (like`wget`), which are able to handle any kind of artifact blob
  with different natural repository types and identity schemes.
- Therefore,
  - it can neither provide an implicit reference hint anymore
  - nor there is a one-to-one relation to a typed uploader anymore.
- artifacts might be uploadable to different repository types using different
  identity schemes.

This problem is partly covered by allowing to specify a hint along with those access methods
similar to the `localBlob` access method. But this can only be a workaround, because

- the hint is not typed and potential target repositories might use different identity schemes
- it is possible to store a single hint, only.

### Proposed Solution

Hints have to be typed to allow uploaders to know what identities are provided, how the
hint string has to be interpreted and to decide for uploaders, which hint to use. Additionally,
it must be possible to store multiple hints for an artifact to support multiple possible upload
repository types.

To be compatible a serialization format is defined for a list of type hints, which maps such
a list to a single string.

The library provides a new type `ReferenceHint`, which provides access to
a formal hint, by providing access to a set of string attributes. The semantic type is a string map
of strings. There are three predefined attributes:

- `type` the formal type of the hint (may be empty to cover old component versions)
  The types are defined like the resource types. The following types are defined for now:
  - `oci`: OCI identity given by the attribute `reference` with the currently used format
  - `maven`: Maven identity (GAV) given by the attribute `reference` with the currently used format
  - `npm`: Node package identity given by the attribute `reference` with the currently used
    format
- `reference`: the standard attribute to hold a string representation for the identity.
- `implicit`: Value `true` indicated an implicit hint (as used today) provided by an accessmethods.

Type-less former hints are represented by the sole attribute `reference`.
New Hint types may use other attributes.

#### Access Methods

An access method can provide (and store) implicit hints as before. Those hints are indicated
to be implicit. When composing an access method it is only allowed to store implicit hints.
This is completely compatible to the current solution.

Additionally, multiple hints can be stored and delivered.

To support any kind of hint for any scenario, the artifact metadata (resources and sources)
is extended to store explicit hints, which will be part of the normalized form.
This is done by an additional attribute `referenceHints`. It is a list of string maps
holding the hint attributes (including the hint type).

#### Uploaders

Uploaders are called with the aggregation of explicit (from metadata) and implicit (from
access methods) hints. Hereby, the explicit hints have precedence.

If an uploader creates a local access specification, only implicit hints may be stored, here.

There is a new composition option (`--refhint`) now for composing resources
and sources for the CLI. It accepts an attribute setting. Multiple such options starting with the `type` attribute are used to compose a single hint.

#### Inputs

Inputs may provide explicit or implicit hints, now. All file based inputs now allow to specify implicit hints as used before.
The implicit hints (if not conflicting with explicit hints) are used to be stored in `localBlob`
access methods. The explicit hints are used to default the explicit artifact hints.

Hints used in a component version must be unique. This check is extended to consider implicit
and explicit hints provided by inputs, access methods and artifact hints.

Hints may either be specified as a list of string maps (canonical form) or as string using the serialized from.

### Hint Serialization Format

In general a hint is serialized to the following string
<center>
    [&lt;*type*>`::]`&lt;*attribute*>`=`&lt;*value*>{`,`&lt;*attribute*>`=`&lt;*value*>}
</center>

The type is not serialized as attribute, but as prefix separated by a `::`. The `implicit` attribute is never serialized if the string is stored in an access specification.
If no type is known the type part is omitted.

A list of hints is serialized to

<center>
   &lt;*hint*>{`;`&lt;*hint*>}
</center>

*Attributes* names consist of alphanumeric characters, only.
A *value* may not contain a `::`. If it contains a `;`, `,` or `"`
character it must be given in double quotes.
In the double-quoted form any `"` or `\` character has to be escaped by
a preceding `\` character.

To be as compatible as possible, a single attribute hint with the attribute
`reference` is serialized as naked value (as before) if there are no special
characters enforcing a quoted form.

### Incompatible Changes

#### Component Version Representation

- The new library creates always typed hints for new elements. Old hints are
  left as they are. This means, that old versions of the OCM tooling
  cannot work correctly with component versions with persisted hints in
  access specifications
- If explicit hints are created, they are not observed by old tool versions.
  Those component versions cannot be verified by an older tool version.

#### OCM Library

- The `SetResourceBlob` and `SetSourceBlob` API methods now accept
  a hint specification instead of a string. To be as compatible as possible,
  it still accepts  a string (as before), which is mapped by a hint parser to a hint list.
  Therefore, the technical type is `interface{}`, which accepts
  various effective types (single hints, string and hint lists).
- Uploaders provided by a plugin now get a serialized hint list
  instead of a simple un-typed reference format.
- There are new options for creating resource(source access objects.
