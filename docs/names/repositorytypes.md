# Repository Types in the Open Component Model

Any repository that can be used to store content according to the
Open Component Model must be describable by a formal repository
specification.

Such a specification is usable by language binding supporting
this kind of specification to gain access to this repository.

Therefore, a repository specification has a type, the *Repository Type*.
Additionally, it defines dedicated attributes, which are
used to determine the access to a dedicated instance of this repository type.

There are two kinds of types:
- centrally defined type names managed by the OCM organization

  The format of a repository type is described by the following regexp:

  ```regex
  [A-Z][a-zA-Z0-9]*
  ```

  The actually defined types with their meaning and format can be
  found [here](../formats/repositories/README.md)
- 
- vendor specific types

  any organization using the open component model may define dedicated types on
  their own. Nevertheless, the meaning of those types must be defined.
  Basically there may be multiple such types provided by different organizations
  with the same meaning. But we strongly encourage organizations to share
  such types instead of introducing new type names.

  To support a unique namespace for those type names vendor specific types 
  have to follow a hierarchical naming scheme based on DNS domain names.
  Every type name has to be suffixed by a DNS domain owned by the providing
  organization (for example `myspecialrepo.acme.com`).
  The local type must follow the above rules for centrally defined type names
  and prepended, separated by a dot (`.`).

  So, the complete pattern looks as follows:

  ```
  [a-z][a-zA-Z0-9].<DNS domain name>
  ```