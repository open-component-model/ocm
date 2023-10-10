# Naming Scheme for Labels in the OCM Component Descriptor

There are several elements in the component descriptor, which
can be annotated by labels:

- The component version itself
- resource specifications
- source specifications
- component version references

Besides the type of an element (for resources and sources), labels
are intended to express additional semantics for an element. 
To do so the meaning of labels must be clearly defied. Therefore,
a label and its bound semantic must be uniquely identified by its name.

The usage of labels is left to the creator of a component version, therefore
the set of labels must be extensible.
Because of this extensibility, the names of labels must be globally
unique, also.

Like for [resource types](resourcetypes.md) there are two flavors
of label names:

- labels with a predefined meaning for the component model itself.

  Those labels are used by the standard OCM library and tool set to
  control some behaviour like signing.

  Such labels use flat names following a camel case scheme with
  the first character in lower case.

  Their format is described by the following regexp:

  ```regex
  [a-z][-a-zA-Z0-9]*
  ```

- vendor specific labels

  any organization using the open component model may define dedicated labels
  on their own. Nevertheless, their names must be globally unique.
  Basically there may be multiple such labels provided by different organizations
  with the same meaning. But we strongly encourage organizations to share
  such types instead of introducing new type names.

  To support a unique namespace for those label names vendor specific labels
  have to follow a hierarchical naming scheme based on DNS domain names.
  Every label name has to be preceded by a DNS domain owned by the providing
  organization (for example `landscaper.gardener.cloud/blueprint`).
  The local name must follow the above rules for centrally defined names
  and is appended, separated by a slash (`/`).

  So, the complete pattern looks as follows:

  ```
  <DNS domain name>/[a-z][-a-zA-Z0-9]*
  ```
  
Every label must define a specification of its attributes,
to describe its value space. This specification may be versioned.
The version must match the following regexp

```
v[0-9]+([a-z][a-z0-9]*)?
```

A label entry in the component descriptor consists of a dedicated set of
attributes with a predefined meaning. While arbitrary values are allowed for the 
label `value`, additional (vendor/user specific) attributes are not
allowed at the label entry level.

- `name` (required) *string*

  The label name according to the specification above.

- `value` (required) *any*

  The label value may be an arbitrary JSON compatible YAML value.

- `version` (optional) *string*

  The specification version for the label content. If no version is
  specified, implicitly the version `v1` is assumed.

- `signing` (optional) *bool*:  (default: `false`)

  If this attribute is set to `true`, the label with its value will be incorporated
  into the signatures of the component version.

  By default, labels are not part of the signature.

Centrally defined labels with their specification versions
can be found [here](../labels/README.md).