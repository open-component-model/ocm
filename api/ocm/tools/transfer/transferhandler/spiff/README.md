# Spiff-based Transfer Handler

This package provides a `TransferHandler` programmable using the
[Spiff Templating Engine](https://github.com/mandelsoft/spiff).
The spiff template get a flat binding `mode` describing the operation mode
and it must provide a top-level node `process` containing
the processing result.

The following modes are used:

## Update Mode

This mode is used to decide on the update option for a component version. This controls whether an update on volatile (non-signature relevant) information should be done. It gets the following bindings:

- `mode` *&lt;string>*: `update`
- `values` *&lt;map>*:
  - `component` *&lt;map>*:  the meta data of the component version carrying the resource
    - `name` *&lt;string>*: component name
    - `version` *&lt;string>*: component version
    - `provider` *&lt;provider>*: provider info, a structure with the fields
      - `name` *&lt;string>*: provider name
      - `labels` *&lt;map[string]any>*: provider attributes
    - `labels` *&lt;map[string]>*: labels of the component version (deep)
  - `target` *&lt;map>*:  the repository specification of the target resource

The result value (field `process`) must be a boolean describing whether the update should be possible.

## Enforce Transport Mode

This mode is used to decide on the enforced transport option for a
component version. This controls whether transport is carried out
as if the component version were not present at the destination.
It gets the following bindings:

- `mode` *&lt;string>*: `enforceTransport`
- `values` *&lt;map>*:
  - `component` *&lt;map>*:  the meta data of the component version carrying the resource
    - `name` *&lt;string>*: component name
    - `version` *&lt;string>*: component version
    - `provider` *&lt;provider>*: provider info, a structure with the fields
      - `name` *&lt;string>*: provider name
      - `labels` *&lt;map[string]any>*: provider attributes
    - `labels` *&lt;map[string]>*: labels of the component version (deep)
  - `target` *&lt;map>*:  the repository specification of the target resource

The result value (field `process`) must be a boolean describing whether the update should be possible.

## Overwrite Mode

This mode is used to decide on the overwrite option for a component version. This controls whether an update on non-volatile (signature relevant) information should be done. It gets the
following bindings:

- `mode` *&lt;string>*: `overwrite`
- `values` *&lt;map>*:
  - `component` *&lt;map>*:  the meta data of the component version carrying the resource
    - `name` *&lt;string>*: component name
    - `version` *&lt;string>*: component version
    - `provider` *&lt;provider>*: provider info, a structure with the fields
      - `name` *&lt;string>*: provider name
      - `labels` *&lt;map[string]any>*: provider attributes
    - `labels` *&lt;map[string]>*: labels of the component version (deep)
  - `target` *&lt;map>*:  the repository specification of the target resource

The result value (field `process`) must be a boolean describing whether the update should be possible.

## Resource Mode

This mode is used to decide on the by-value option for a resource. It gets the
following bindings:

- `mode` *&lt;string>*: `resource`
- `values` *&lt;map>*:
  - `component` *&lt;map>*:  the meta data of the component version carrying the resource
    - `name` *&lt;string>*: component name
    - `version` *&lt;string>*: component version
    - `provider` *&lt;provider>*: provider info, a structure with the fields
      - `name` *&lt;string>*: provider name
      - `labels` *&lt;map[string]any>*: provider attributes
    - `labels` *&lt;map[string]>*: labels of the component version (deep)
  - `element` *&lt;map>*:  the meta data of the resource and the field `type` containing the element type.
  - `access` *&lt;map>*:  the access specification of the resource
  - `target` *&lt;map>*:  the repository specification of the target resource

The result value (field `process`) must be a boolean describing whether the
resource should be transported ny-value.

## Source Mode

This mode is used to decide on the by-value option for a source. It gets the
following bindings:

- `mode` **&lt;string>**: `source`
- `values` **&lt;map>**: (see [Resource Mode](#resource-mode))

The result value (field `process`) must be a boolean describing whether the
resource should be transported ny-value.

## Component Version Mode

This mode is used to decide on the recursion option for a referenced component
version. It gets the  following bindings:

- `mode` **&lt;string>**: `componentversion`
- `values` **&lt;map>**: (see [Resource Mode](#resource-mode))
  - `component` **&lt;map>** *(optional)*:  the meta data of the component version carrying the reference
    - `name` **&lt;string>**: component name
    - `version` **&lt;string>**: component version
    - `provider` **&lt;string>**: provider name
    - `labels` **&lt;map[string]>**: labels of the component version (deep)
  - `element` **&lt;map>**:  the meta data of the component reference

The result value (field `process`) can either be a simple boolean value
or a map with the following fields:

- `process` **&lt;bool>**: `true` indicates to follow the reference
- `repospec` **&lt;map>** *(optional)*: the specification of the repository to use to follow the reference
- `script` **&lt;template>** *(optional)*: the script to use instead of the current one.

If no new repository spec is given, the actual repository is used. If no new
script is given, the actual one is used for sub sequent processing.

the `component` field is optional, for top level requests during the transport
of a set of component versions, there is no hosting component.
