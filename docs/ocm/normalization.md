# Component Descriptor Normalization

The [component descriptor](../formats/compdesc/README.md) is used to describe
a [component version](model.md#component-versions). It contains several kinds
of information:
- volatile label settings, which might be changeable.
- artifact access information, which might be changed during transport steps.
- static information describing the features and artifacts of a component
  version.

<!-- If a component version should be signed, to be able to verify its authenticity
after transportation steps, the technical representation of a component descriptor
cannot be used to calculate the digest, which is finally signed. Only the last
kind of content must be covered by the signature, because ethe other information
might be changed over time. -->

For signing a digest of the component descriptor needs to be generated.
Therefore, a standardized normalized form is needed, which contains only the signature relevant
information. This is the source to calculate a digest, which is finally signed (and verified).

Like for signature algorithms, the model offers the possibility to work with
different normalization algorithms/formats.

To support legacy versions of the component model, there are two different
normalizations.
- `JsonNormalisationV1`: This is a legacy format, which depends of the format of the
  component descriptor
- `JsonNormalisationV2`: This is the new format. which is independent of the
  chosen representation format of the component desriptor.

## Generic Normalization format

The generic format is based on a data structure consisting of dictionaries, lists and
simple values (like strings and integers).

The signing relevant information described by a component descriptor is mapped
to such a data structure according to the format specifications described below.

This data structure is then mapped to a formal JSON representation, which
only contains clearly ordered elements. It is marshalled without white-spaces contained
in the representation. Therefor, the resulting byte stream is directly defined
by the inbound data structure and independent of the order of marshalling
dictionaries/objects.
Its digest can be used as basis to calculate a signature.

To map lists and dictionaries into such clearly ordered elements the rules described
below are used. The inbound data structures in the examples below are shown in
YAML notation.

### Simple Values

Simple values are kept as they are.

Example:
```yaml
  "bob"
```
will result in :

```json
  "bob"
```
### Dictionary

All dictionaries are converted to a list where each element is a single-entry
dictionary containing the key/value pair of the original entry. This list is
ordered by lexicographical order of the keys.

Example:
```yaml
  bob: 26
  alice: 25
```
will result in :

```json
  [{"alice":25},{"bob":26}]
```

The values are converted according to the same rules, recursively.

Example:
```yaml
  people:
    bob: 26
    alice: 25
```
will result in :

```json
  [{"people":[{"alice":25},{"bob":26}]}]
```

### Lists

Lists are converted to JSON arrays and preserve the order of the elements

Example:
```yaml
- bob
- alice
```

normalized to:
```json
["bob","alice"]
```

The values are converted according to the same rules, recursively.

Example:
```yaml
   - bob: 26
   - alice: 25
```

will result in :

```json
  [[{"bob":26}],[{"alice":25}]]
```

### Combined example

The following snippet is taken from a real component descriptor.

```yaml
resources:
- access:
    localReference: blob
    mediaType: text/plain
    referenceName: ref
    type: localBlob
  extraIdentity:
    additional: value
    other: othervalue
  name: elem1
  relation: local
  type: elemtype
  version: 1
```

normalized to

```json
[{"resources":[[{"access":[{"localReference":"blob"},{"mediaType":"text/plain"},{"referenceName":"ref"},{"type":"localBlob"}]},{"extraIdentity":[{"additional":"value"},{"other":"othervalue"}]},{"name":"elem1"},{"relation":"local"},{"type":"elemtype"},{"version":1}]]}]
```

formatted with white spaces for better readability it looks like:

```json
[
  {
    "resources": [
      [
        {
          "access": [
            {
              "localReference": "blob"
            },
            {
              "mediaType": "text/plain"
            },
            {
              "referenceName": "ref"
            },
            {
              "type": "localBlob"
            }
          ]
        },
        {
          "extraIdentity": [
            {
              "additional": "value"
            },
            {
              "other": "othervalue"
            }
          ]
        },
        {
          "name": "elem1"
        },
        {
          "relation": "local"
        },
        {
          "type": "elemtype"
        },
        {
          "version": 1
        }
      ]
    ]
  }
]
```

### Empty values:

Empty lists are normalized as empty lists

```yaml
myList: []
```

```json
[{"myList":[]}]
```

Null values are skipped during initialization

```yaml
myList: ~
```

```yaml
myList: null
```

```yaml
myList:
```
are all normalized to:

```json
[]
```

## Relevant information in Component Descriptors.

A component descriptor contains signature relevant information and
information, which may change. For example, the access methods specifications
might be changed during atransport step.

Relevant fields and their mapping to the normalized data structure for `JsonNormalisationV2`:
- Component Name: mapped to `component.name`
- Component Version: mapped to `component.version`
- Component Labels: mapped to `component.labels` (see [Labels](#labels)])
- Component Provider: mapped to `component.provider`
- Resources: mapped to `component.resources`, always empty list enforced, without the source references (see [Labels](#labels)] and [Access Methods](#access-methods)])
- Sources: mapped to `component.sources`, always empty list enforced, (see [Labels](#labels)] and [Access Methods](#access-methods)])
- References: mapped to `component.references`, always empty list enforced, (see [Labels](#labels)])


### Access Methods

Access method specifications are completely ignored.
A resource/source is ignored, if the access method type is `none`.

## Labels

Labels are removed before signing but can be marked with a special boolean
property `signing` not to be removed and thus be part of the signature.
The structure of signing-relevant labels is preserved from the component
descriptor version `v2`.

Example:

```yaml
labels:
- name: label1
  value: foo
- name: label2
  value: bar
  signing: true
```

`label1` will be excluded from the signature, `label2` will be included.
The label values is takes as it is, preserving a potentially deep structure.


# `JsonNormalisationV1` vs `JsonNormalisationV2`

The `JsonNormalisationV1` serialization format is based on the serialization
format of the component descriptor. The format version fields are included

`JsonNormalisationV2` strictly uses only the relevant component descriptor
information according to the field specification shown above.
