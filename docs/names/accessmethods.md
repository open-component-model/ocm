# Naming Scheme for Access Methods

Access methods describe (and finally implement) dedicated technical ways how to
access the blob content of a (re)source described by an
[OCM component descriptor](../formats/compdesc/README.md).

They are an integral part of the Open Component Model. They always
provide the link between a component version stored in some repository context,
and the technical access of the described resources applicable for this
context. Therefore, the access method of a resource may change when
component versions are transported among repository contexts.

In a dedicated context all used access methods must be known by the used tool
set. Nevertheless, the set of access methods is not fixed. The actual
library/tool version provides a simple way to locally add new methods with
their implementations to support own local environments.

Because of this extensibility, the names of access methods must be globally
unique.

Like for [resource types](resourcetypes.md), there are two flavors
of method names:

- centrally provided access methods

  Those methods are coming with the standard OCM library and tool set.
  It provides an implementation and component version using only such
  access methods can be use across local organizational extension.

  These types use flat names following a camel case scheme with
  the first character in lower case (for example `ociArtefact`).

  Their format is described by the following regexp:

  ```regex
  [a-z][a-zA-Z0-9]*
  ```

- vendor specific types

  any organization using the open component model may define dedicated access
  methods on their own. Nevertheless, their name must be globally unique.
  Basically there may be multiple such types provided by different organizations
  with the same meaning. But we strongly encourage organizations to share
  such types instead of introducing new type names.

  Extending the toolset by own access methods always means to locally
  provide a new tool version with the additionally registered access method
  implementations. Because the purpose of the Open Component Model is the
  exchange of software, the involved parties must agree on the used toolset.
  This might involve methods provided by several potentially non-central 
  providers. Therefore, use used access method names must be globally unique
  with a globally unique meaning.

  To support a unique namespace for those type names vendor specific types
  have to follow a hierarchical naming scheme based on DNS domain names.
  Every type name has to be suffixed by a DNS domain owned by the providing
  organization.
  The local type must follow the above rules for centrally defined type names
  and suffixed by the namespace separated by a dot (`.`)

  So, the complete pattern looks as follows:

  ```
  [a-z][a-zA-Z0-9]*\.<DNS domain name>
  ```
  
Every access method type must define a specification of its attributes,
required to locate the content. This specification may be versioned.
Therefore, the type name used in an access specification in the component descriptor
may include a specification version appended by a slash (`/`).
Similar to the kubernetes api group versions, the version must match the
following regexp

```
v[0-9]+([a-z][a-z0-9]*)?
```

Examples:
- `ociArtefact/v1`
- `myprotocol.acme.org/v1alpha1`

If no version is specified, implicitly the version `v1` is assumed.

Centrally defined access methods with their specification versions
can be found [here](../formats/accessmethods/README.md).