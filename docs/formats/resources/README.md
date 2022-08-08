# Open Component Model Resource Types

The following resource types are centrally defined:

- `ociArtefact`: a generic OCI artefact follwoing the
   [open containers image specification](https://github.com/opencontainers/image-spec/blob/main/spec.md)
- `ociImage`: an OCI image or image list
- `helmChart`: a helm chart, either stored as OCI artefact or as tar blob (tar media type)
- `blob`: any anonymous untyped blob data
- `filesytem`: a directory structures stored as archive (tar, tgz).

For centrally defined resource types, there might be special support in the
OCM library and tool set. For example, there is a dedicated downloader
for helm charts providing the filesystem helm chart format regardless of
the storage method and supported media type.

Besides those types, there are some vendor types that are typically used:

- `landscaper.gardener.cloud/blueprint`: an installation description for the landscaper tool
- `landscaper.gardener.cloud/gitOpsTemplate`: a filesystem content (tar, tgz)
  intended to be used as git ops template to setup a git repo used for continuous deployment (for example flux)