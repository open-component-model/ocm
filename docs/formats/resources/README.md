# Open Component Model Resource Types

The following resource types are centrally defined:

- `ociArtefact`
- `ociImage`
- `helmChart`
- `blob`
- `filesytemContent`

For centrally defined resource types there might be special support in the
OCM library and tool set. For example there is a dedicated downloader
for hem charts providing the filesystem helm chart format regardless of
the storage method.

Besides those types, there are some vendor types that are typically used:

- landscaper.gardener.cloud/blueprint
- landscaper.gardener.cloud/gitOpsTemplate