# Localization Tools

This package (pkg/contexts/ocm/util/localize) contains some
yaml format definitions, (format.go) given by a Go structure.
and functions usable for the modification of filesystem snapshots.

It covers
- OCM based localization descriptions used to describe image 
  location substitution and the
- merging with configuration input.

This mechanism is intended to support the mapping of generic filesystem
snapshots containing deployment descriptions or manifests to
an installation specific snapshot incorporating the local location
of container images based on OCM component versions and installation specific
configuration values.

The description format describes two basic specifications that incorporate external 
information provided by a component version or some user config.

- struct/format `Localization` describe a specification to
  inject/modify image locations based on the information provided
  by a component version. The substitution descriptions use relative resources
  references to specify the resource whose image reference is used as basis
  for the substituted value.

  This specification is intended to be stored as part of a resource artifact in a
  component version which is then used to apply it. Thereby the contained relative
  [resource reference](../../../../../docs/ocm/model.md#resource-reference)
  are evaluated against the component version containing the specification to resolve
  the final image location.

- struct/format `Configuration` describes a specification for
  applying a dedicated config value taken from a configuration source
  to a filesytem snapshot.

The function `Localize` and `Configure` accept a list of such 
specifications and map them into an environment agnostic set of
`Substitution` specifications, which contain resolved data values, only.
A third function `Substitute` takes those environment agnostic specifications
and apply them to a filesystem.

Finally, a compound specification `InstantiationRules` is provided,
that combines all those descriptions with the specification of the snapshot
resource and further helper parts, like json scheme validation for config files.

Such a specification object can be applied by the function `Instantiate` 
together with configuration values to
a component version. As substitution result it returns a virtual filesystem
with the snapshot according to the resolved substitutions. To get access to the
template resource containing the filesystem snapshot to be instantiated, the
configured downloaders (package `pkg/context/ocm/download`) is used.
Therefore, this method can be used together with any own resource type as long as 
an appropriate downloader is configured.

Additionally, there is a set of more basic types and methods, which can be used
to describe end execute localizations for single data objects (see `ImageMappings`,
`LocalizeMappings` and `SubstituteMappings`).