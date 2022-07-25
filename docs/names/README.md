# Naming Schemes used in the Open Component Model

In OCM there are several kinds of type names. For all those kinds there
is a dedicated naming scheme.

- [Label Names](labels.md)

  The OCM component descriptor itself, resources, sources  and component version
  references can be enriched by labels capable to carry values with an
  arbitrary structure.

- [Resource Type Names](resourcetypes.md)

  The OCM component descriptor describes a set of resources, their type and
  meaning with attached meta and access information.
  
- [Access Method Names](accessmethods.md)

  Access methods describe dedicated technical ways how to access the blob
  content of a (re)source described by an 
  [OCM component descriptor](../formats/compdesc/README.md). It is evaluated in
  the storage context used to read the component descriptor containing the
  access method description.