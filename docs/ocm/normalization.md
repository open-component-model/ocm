# Component Descriptor Normalization

The [component descriptor](../formats/compdesc/README.md) is used to describe
a [component version](model.md#component-versions). It contains several kinds
of information:
- volatile label settings, which might be changeable.
- artifact access information, which might be changed during transport steps.
- static information describing the features and artifacts of a component 
  version.

If a component version should be signed, to be able to verify its authenticity
after transportation steps, the technical representation of a component descriptor
cannot be used to calculate the digest, which is finally signed. Only the last
kind of content must be covered by the signature, because ethe other information
might be changed over time.

Therefore, a standardized normalized form of a component descriptor is generated,
which contains only the signature relevant information. This is then the technical
basis to calculate a digest, which is finally signed (and verified).

Like for signature algorithms, the model offers the possibility to work with
different normalization algorithms/formats.

To support legacy versions of the component model, there are two different
normalizations.
- `JsonNormalisationV1`: This is a legacy format, which depends of the format of the
  component descriptor
- `JsonNormalisationV2`: This is the new format. which is independent of the 
  chosen representation format of the component desriptor.

## Generic Normalization format

The normalization of a component descriptor is based of a generic
ordered JSON representation of the normalized data structure.


