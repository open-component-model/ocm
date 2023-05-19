## ocm ocm-downloadhandlers &mdash; List Of All Available Download Handlers

### Description


A download handler can be used to process resources to be downloaded from
on OCM repository. By default, the blobs provided from the access method
(see [ocm ocm-accessmethods](ocm_ocm-accessmethods.md)) are used to store the resource content
in the local filesystem. Download handlers can be used to tweak this process.
They get access to the blob content and decide on their own what to do
with it, or how to transform it into files stored in the file system.

For example, a pre-registered helm download handler will store
OCI-based helm artifacts as regular helm archives in the local
file system.

### Handler Registration 

Programmatically any kind of handlers can be registered for various
download conditions. But this feature is available as command-line option, also.
New handlers can be provided by plugins. In general available handlers,
plugin-based or as part of the CLI coding are nameable using an hierarchical
namespace. Those names can be used by a <code>--downloader</code> option
to register handlers for various conditions for CLI commands like
[ocm download resources](ocm_download_resources.md) (implicitly registered download handlers
can be enabled using the option <code>-d</code>).

Besides the activation constraints (resource type and media type of the
resource blob), it is possible to pass handler configuration controlling the
exact behaviour of the handler for selected artifacts.

The following handler names are possible:
  - <code>plugin</code>: [downloaders provided by plugins]
    
    sub namespace of the form <code>&lt;plugin name>/&lt;handler></code>

  - <code>ocm/dirtree</code>: downloading directory tree-like resources
    
    The <code>dirtree</code> downloader is able to to download directory-tree like
    resources as directory stricture (default) or archive.
    The following artifact media types are supported:
      - <code>application/vnd.oci.image.manifest.v1+tar+gzip</code>
      - <code>application/x-tgz</code>
      - <code>application/x-tar+gzip</code>
      - <code>application/x-tar</code>
    
    By default, it is registered for the following resource types:
      - <code>directoryTree</code>
      - <code>filesystem</code>
    
    If accepts a config with the following fields:
      - <code>asArchive</code>: flag to request an archive download
      - <code>configTypes</code>: a list of accepted OCI config archive types
        defaulted by <code>application/vnd.oci.image.config.v1+json</code>.



See [ocm ocm-downloadhandlers](ocm_ocm-downloadhandlers.md) for further details on using
download handlers.


### SEE ALSO

##### Parents

* [ocm](ocm.md)	 &mdash; Open Component Model command line client



##### Additional Links

* [<b>ocm ocm-accessmethods</b>](ocm_ocm-accessmethods.md)	 &mdash; List of all supported access methods
* [<b>ocm download resources</b>](ocm_download_resources.md)	 &mdash; download resources of a component version
* [<b>ocm ocm-downloadhandlers</b>](ocm_ocm-downloadhandlers.md)	 &mdash; List of all available download handlers

