# Centrally Artifact Normalization Algorithms


To be able to sign a component version, the content of described artifacts
must be incorporated. Therefore, a digest for the artifact content must be
determined.

By default, this digest is calculated based on the blob provided by the 
[access method](../../ocm/model.md#artifact-access)
of an artifact. But there might be technology specific ways to uniquely identify
the content for dedicated artifact types.

Therefore, together with the digest and its algorithm, an artifact normalization
algorithm is kept in the [component descriptor](../../ocm/model.md#component-descriptor).

The following algorithms are centrally defined and available in the OCM toolset:

- `ociArtifactDigest/v1`: OCI manifest digest

  This algorithm is used for artifacts of type `ociArtifact`. It just uses the
  manifest digest of the OCI artifact.

- `genericBlobDigest/v1`: Blob byte stream digest
  
  This is the default normalization algorithm. It just uses the blob content
  provided by the access method of an OCM artifact to calculate the digest. 
  It is always used, if no special digester is available for an artifact type.