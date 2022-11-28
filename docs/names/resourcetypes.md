# Resource Types in the Open Component Model

The OCM component descriptor described a set of artifacts, their type and
meaning with attached meta, and access information.

The meaning is basically encoded into a dedicated *resource type*.
Therefore, the resource type must be globally unique.
The OCM defines a dedicated naming scheme to guarantee this uniqueness.

There are two kinds of types:
- centrally defined type names managed by the OCM organization

  These types use flat names following a camel case scheme with
  the first character in lower case (for example `ociArtifact`).

  Their format is described by the following regexp:
  
  ```regex
  [a-z][a-zA-Z0-9]*
  ```
  
  The actually defined types with their meaning and format can be
  found [here](../formats/resources/README.md)
  
- vendor specific types

  any organization using the open component model may define dedicated types on
  their own. Nevertheless, the meaning of those types must be defined.
  Basically there may be multiple such types provided by different organizations
  with the same meaning. But we strongly encourage organizations to share
  such types instead of introducing new type names.

  To support a unique namespace for those type names vendor specific types
  have to follow a hierarchical naming scheme based on DNS domain names.
  Every type name has to be preceded by a DNS domain owned by the providing
  organization (for example `landscaper.gardener.cloud/blueprint`).
  The local type must follow the above rules for centrally defined type names 
  and is appended, separated by a slash (`/`).

  So, the complete pattern looks as follows:

  ```
  <DNS domain name>/[a-z][a-zA-Z0-9]*
  ```