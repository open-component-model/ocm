---
title: "ocm ocm-uploadhandlers &mdash; List Of All Available Upload Handlers"
linkTitle: "ocm-uploadhandlers"
url: "/docs/cli-reference/ocm-uploadhandlers/"
sidebar:
  collapsed: true
---

### Description

An upload handler is used to process resources using the access method
<code>localBlob</code> transferred into an OCM
repository. They may decide to store the content in some other
storage repository. This may be an additional storage location or it
may replace the storage of the resource as local blob.
If an additional storage location is chosen, the local access method
is kept and the additional location can be registered in the component
descriptor as <code>globalAccess</code> attribute of the local access
specification.

For example, there is a default upload handler responsible for OCI artifact
blobs, which provides regular OCI artifacts for a local blob, if
the target OCM repository is based on an OCI registry. Hereby, the
<code>referenceName</code> attribute will be used to calculate a
meaningful OCI repository name based on the repository prefix
of the OCM repository (parallel to <code>component-descriptors</code> prefix
used to store the component descriptor artifacts).

### Handler Registration

Programmatically any kind of handlers can be registered for various
upload conditions. But this feature is available as command-line option, also.
New handlers can be provided by plugins. In general available handlers,
plugin-based or as part of the CLI coding are nameable using an hierarchical
namespace. Those names can be used by a <code>--uploader</code> option
to register handlers for various conditions for CLI commands like
[ocm transfer componentversions](ocm_transfer_componentversions.md) or [ocm transfer commontransportarchive](ocm_transfer_commontransportarchive.md).

Besides the activation constraints (resource type and media type of the
resource blob), it is possible to pass a target configuration controlling the
exact behaviour of the handler for selected artifacts.

The following handler names are possible:
  - <code>ocm/mavenPackage</code>: uploading maven artifacts

    The <code>ocm/mavenPackage</code> uploader is able to upload maven artifacts (whole GAV only!)
    as artifact archive according to the maven artifact spec.
    If registered the default mime type is: application/x-tgz

    It accepts a plain string for the URL or a config with the following field:
    'url': the URL of the maven repository.

  - <code>ocm/npmPackage</code>: uploading npm artifacts

    The <code>ocm/npmPackage</code> uploader is able to upload npm artifacts
    as artifact archive according to the npm package spec.
    If registered the default mime type is: application/x-tgz

    It accepts a plain string for the URL or a config with the following field:
    'url': the URL of the npm repository.

  - <code>ocm/ociArtifacts</code>: downloading OCI artifacts

    The <code>ociArtifacts</code> downloader is able to download OCI artifacts
    as artifact archive according to the OCI distribution spec.
    The following artifact media types are supported:
      - <code>application/vnd.oci.image.manifest.v1+tar</code>
      - <code>application/vnd.oci.image.manifest.v1+tar+gzip</code>
      - <code>application/vnd.oci.image.index.v1+tar</code>
      - <code>application/vnd.oci.image.index.v1+tar+gzip</code>
      - <code>application/vnd.docker.distribution.manifest.v2+tar</code>
      - <code>application/vnd.docker.distribution.manifest.v2+tar+gzip</code>
      - <code>application/vnd.docker.distribution.manifest.list.v2+tar</code>
      - <code>application/vnd.docker.distribution.manifest.list.v2+tar+gzip</code>

    By default, it is registered for these mimetypes.

    It accepts a config with the following fields:
      - <code>namespacePrefix</code>: a namespace prefix used for the uploaded artifacts
      - <code>ociRef</code>: an OCI repository reference
      - <code>repository</code>: an OCI repository specification for the target OCI registry

    Alternatively, a single string value can be given representing an OCI repository
    reference.

  - <code>plugin</code>: [downloaders provided by plugins]

    sub namespace of the form <code>&lt;plugin name>/&lt;handler></code>



See [ocm ocm-uploadhandlers](ocm_ocm-uploadhandlers.md) for further details on using
upload handlers.

### SEE ALSO

#### Parents

* [ocm](ocm.md)	 &mdash; Open Component Model command line client



##### Additional Links

* [<b>ocm transfer componentversions</b>](ocm_transfer_componentversions.md)	 &mdash; transfer component version
* [<b>ocm transfer commontransportarchive</b>](ocm_transfer_commontransportarchive.md)	 &mdash; transfer transport archive
* [<b>ocm ocm-uploadhandlers</b>](ocm_ocm-uploadhandlers.md)	 &mdash; List of all available upload handlers

