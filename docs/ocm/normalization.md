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

The normalization of a component descriptor consists of a JSON formatted representation of the
component descriptor containing the relevant date and preserving the order of fields. The resulting
JSON string will then be

### Maps
All maps are converted to a list where each element is a single-entry dictionary containing the key,
value pair of the original entry. The list is ordered in lexikographic order of the keys.

Example:
```
  component:
    name: github.com/vasu1124/introspect
    version: 1.0.0
    provider: internal
```
will result in (for better readability formatted, not the format which is signed):

```
[
    {
        "component": [
            {
                "name": "github.com/vasu1124/introspect"
            },
            {
                "provider": "name": "internal"
            },
            {
                "version": "1.0.0"
            }
        ]
    }
]
```

### Lists
Lists are converted to JSON arrays and preserve the order of the elements

Example:
```
myList:
- foo
- bar
- baz
```

normalized to (for better readability formatted, not the format which is signed):
```
[
  {
    "list": [
      "foo",
      "bar",
      "baz"
    ]
  }
]
```

### Combined example

```
myList:
  - foo
  - bar
  - some: thing
    hello: 26
```

normalized to (for better readability formatted, not the format which is signed):

```
[
  {
    "myList": [
      "foo",
      "bar",
      [
        {
          "hello": 26
        }
       ],
      [
        {
          "some": "thing"
        }
      ]
    ]
  }
]
```

### Empty values:

Empty lists are normalized as empty lists

```
myList: []
```

```
[
  {
    "myList": []
  }
]
```

Null values are skipped during initialization

```
myList: ~
```
```
myList: null
```
```
myList:
```
are all normalized to:

```
[
]
```

## Excluded elements

The following elements are removed during normalization

* meta
* component/repositoryContext
* resources/access
* resources/srcRef
* resources/labels (unless marked for signing)
* sources/access
* sources/labels (unless marked for signing)
* references/labels (unless marked for signing)
* signatures

## Labels
Labels are removed before signing but can be marked with a special boolean property `signing` not to
be removed and thus be part of the signature.

Example:

```
labels:
- name: label1
  value: foo
- name: label2
  value: bar
  signing: true
```
label1 will be excluded from the signature, label2 will be included.


# Differences between JsonNormalisationV1 and JsonNormalisationV2

The JsonNormalisationV1 includes the meta tag of the component descriptor, JsonNormalisationV2
ignores the meta field and only uses the component part.