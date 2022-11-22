# Centrally Artefact Normalization Algorithms


To be able to sign a component version, the content of described artefacts
must be incorporated. Therefore, a digest for the artefact content must be
determined.

By default, this digest is calculated based on the blob provided by the 
[access method](../../ocm/model.md#artefact-access)
of an artefact. But there might be technology specific ways to uniquely identify
the content for dedicated artefact types.

Therefore, together with the digest and its algorithm, an artefact normalization
algorithm is kept in the [component descriptor](../../ocm/model.md#component-descriptor).

The following algorithms are centrally defined and available in the OCM toolset:

- `ociArtefactDigest/v1`: OCI manifest digest

  This algorithm is used for artefacts of type `ociArtefact`. It just uses the
  manifest digest of the OCI artefact.

- `genericBlobDigest/v1`: Blob byte stream digest
  
  This is the default normalization algorithm. It just uses the blob content
  provided by the access method of an OCM artefact to calculate the digest. 
  It is always used, if no special digester is available for an artefact type.