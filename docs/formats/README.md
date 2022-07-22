## Format Specifications

The Open Component Model (OCM) uses some dedicated formats 
to represent content:

- File Formats

  - Common Transport Format (CTF)
  
    It is possible to represent OCI and OCM content as file system content.
    This is used to provide a repository implementation based on a filesystem, 
    which can be used to transport content without direct internet access.

    There are three different technical flavors:
    - `directory`: the content is store directly as a directory tree
    - `tar`: the directory tree is stored in a tar archive
    - `tgz`: the directory tree is stored in a zipped tar archive

    All those technical representations use the same [file formats and directory
    structure](../../pkg/contexts/oci/repositories/ctf/README.md). 

  - Raw Helm Chart Format
  
    Helm Charts can be stored as OCI artefacts in OCI repositories. When
    downloading those artefacts by default they will be transformed to the
    regular helm file system representation by some download handlers.
    If a raw format is chosen the 
    [OCI representation](../../pkg/contexts/oci/repositories/ctf/README.md#artefact-set-archive-format)
    is used to represent the chart content.
  
- Component Descriptor
  
  There are two [serialization versions](compdesc/README.md) to store an OCM component descriptor
  as YAML