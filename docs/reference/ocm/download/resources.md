---
title: "ocm download resources &mdash; Download Resources Of A Component Version"
linkTitle: "download resources"
url: "/docs/cli-reference/download/resources/"
sidebar:
  collapsed: true
---

### Synopsis

```bash
ocm download resources [<options>]  <component> {<name> { <key>=<value> }}
```

#### Aliases

```text
resources, resource, res, r
```

### Options

```text
      --check-verified              enable verification store
  -c, --constraints constraints     version constraint
  -d, --download-handlers           use download handler if possible
      --downloader <name>=<value>   artifact downloader (<name>[:<artifact type>[:<media type>[:<priority>]]]=<JSON target config>) (default [])
  -x, --executable                  download executable for local platform
  -h, --help                        help for resources
      --latest                      restrict component versions to latest
      --lookup stringArray          repository name or spec for closure lookup fallback
  -O, --outfile string              output file or directory
  -r, --recursive                   follow component reference nesting
      --repo string                 repository name or spec
  -t, --type stringArray            resource type filter
      --verified string             file used to remember verifications for downloads (default "~/.ocm/verified")
      --verify                      verify downloads
```

### Description

Download resources of a component version. Resources are specified
by identities. An identity consists of
a name argument followed by optional <code>&lt;key>=&lt;value></code>
arguments.

The option <code>-O</code> is used to declare the output destination.
For a single resource to download, this is the file written for the
resource blob. If multiple resources are selected, a directory structure
is written into the given directory for every involved component version
as follows:

<center>
    <pre>&lt;component>/&lt;version>{/&lt;nested component>/&lt;version>}</pre>
</center>

The resource files are named according to the resource identity in the
component descriptor. If this identity is just the resource name, this name
is used. If additional identity attributes are required, this name is
append by a comma separated list of <code>&lt;name>=&lt;>value></code> pairs
separated by a "-" from the plain name. This attribute list is alphabetical
order:

<center>
    <pre>&lt;resource name>[-[&lt;name>=&lt;>value>]{,&lt;name>=&lt;>value>}]</pre>
</center>



If the option <code>--constraints</code> is given, and no version is specified
for a component, only versions matching the given version constraints
(semver https://github.com/Masterminds/semver) are selected.
With <code>--latest</code> only
the latest matching versions will be selected.


If the <code>--repo</code> option is specified, the given names are interpreted
relative to the specified repository using the syntax

<center>
    <pre>&lt;component>[:&lt;version>]</pre>
</center>

If no <code>--repo</code> option is specified the given names are interpreted
as located OCM component version references:

<center>
    <pre>[&lt;repo type>::]&lt;host>[:&lt;port>][/&lt;base path>]//&lt;component>[:&lt;version>]</pre>
</center>

Additionally there is a variant to denote common transport archives
and general repository specifications

<center>
    <pre>[&lt;repo type>::]&lt;filepath>|&lt;spec json>[//&lt;component>[:&lt;version>]]</pre>
</center>

The <code>--repo</code> option takes an OCM repository specification:

<center>
    <pre>[&lt;repo type>::]&lt;configured name>|&lt;file path>|&lt;spec json></pre>
</center>

For the *Common Transport Format* the types <code>directory</code>,
<code>tar</code> or <code>tgz</code> is possible.

Using the JSON variant any repository types supported by the
linked library can be used:

OCI Repository types (using standard component repository to OCI mapping):

  - <code>CommonTransportFormat</code>: v1
  - <code>OCIRegistry</code>: v1
  - <code>oci</code>: v1
  - <code>ociRegistry</code>



If the <code>--downloader</code> option is specified, appropriate downloader handlers
are configured for the operation. It has the following format

<center>
    <pre>&lt;name>:&lt;artifact type>:&lt;media type>=&lt;yaml target config></pre>
</center>

The downloader name may be a path expression with the following possibilities:
  - <code>helm/artifact</code>: download helm chart
    resources

    The <code>helm</code> downloader is able to download helm chart resources as
    helm chart packages. Thus, the downloader may perform transformations.
    For example, if the helm chart is currently stored as an oci artifact, the
    downloader performs the necessary extraction to provide the helm chart package
    from within that oci artifact.

    The following artifact media types are supported:
      - <code>application/vnd.oci.image.manifest.v1+tar+gzip</code>
      - <code>application/vnd.cncf.helm.chart.content.v1.tar+gzip</code>

    It accepts no config.

  - <code>landscaper/blueprint</code>: uploading an OCI artifact to an OCI registry

    The <code>artifact</code> downloader is able to transfer OCI artifact-like resources
    into an OCI registry given by the combination of the download target and the
    registration config.

    If no config is given, the target must be an OCI reference with a potentially
    omitted repository. The repo part is derived from the reference hint provided
    by the resource's access specification.

    If the config is given, the target is used as repository name prefixed with an
    optional repository prefix given by the configuration.

    The following artifact media types are supported:
      - <code>application/vnd.docker.distribution.manifest.v2+tar</code>
      - <code>application/vnd.docker.distribution.manifest.v2+tar+gzip</code>
      - <code>application/vnd.gardener.landscaper.blueprint.layer.v1.tar</code>
      - <code>application/vnd.gardener.landscaper.blueprint.layer.v1.tar+gzip</code>
      - <code>application/vnd.gardener.landscaper.blueprint.v1+tar</code>
      - <code>application/vnd.gardener.landscaper.blueprint.v1+tar+gzip</code>
      - <code>application/vnd.oci.image.manifest.v1+tar</code>
      - <code>application/vnd.oci.image.manifest.v1+tar+gzip</code>
      - <code>application/x-tar</code>
      - <code>application/x-tar+gzip</code>
      - <code>application/x-tgz</code>

    It accepts a config with the following fields:
      - <code>ociConfigTypes</code>: a list of accepted OCI config archive mime types
        defaulted by <code>application/vnd.gardener.landscaper.blueprint.config.v1</code>.



    This handler is by default registered for the following artifact types:
    landscaper.gardener.cloud/blueprint,blueprint

  - <code>oci/artifact</code>: downloading an OCI artifact
    and optionally re-uploading to an OCI registry

    The <code>artifact</code> download resources stored as oci artifact.
    Furthermore, it allows to specify another OCI registry as download destination,
    thereby, providing a kind of transfer functionality.

    If no config is given, the target must be an OCI reference with a potentially
    omitted repository. The repo part is derived from the reference hint provided
    by the resource's access specification.

    If the config is given, the target is used as repository name prefixed with an
    optional repository prefix given by the configuration.

    The following artifact media types are supported:
      - <code>application/vnd.oci.image.manifest.v1+tar+gzip</code>
      - <code>application/vnd.oci.image.index.v1+tar+gzip</code>

    It accepts a config with the following fields:
      - <code>namespacePrefix</code>: a namespace prefix used for the uploaded artifacts
      - <code>ociRef</code>: an OCI repository reference
      - <code>repository</code>: an OCI repository specification for the target OCI registry

  - <code>ocm/dirtree</code>: downloading directory tree-like resources

    The <code>dirtree</code> downloader is able to download directory-tree like
    resources as directory structure (default) or archive.
    The following artifact media types are supported:
      - <code>application/vnd.oci.image.manifest.v1+tar+gzip</code>
      - <code>application/x-tgz</code>
      - <code>application/x-tar+gzip</code>
      - <code>application/x-tar</code>

    By default, it is registered for the following resource types:
      - <code>directoryTree</code>
      - <code>filesystem</code>

    It accepts a config with the following fields:
      - <code>asArchive</code>: flag to request an archive download
      - <code>ociConfigTypes</code>: a list of accepted OCI config archive mime types
        defaulted by <code>application/vnd.oci.image.config.v1+json</code>.

  - <code>plugin</code>: [downloaders provided by plugins]

    sub namespace of the form <code>&lt;plugin name>/&lt;handler></code>



See [ocm ocm-downloadhandlers](ocm_ocm-downloadhandlers.md) for further details on using
download handlers.



The library supports some downloads with semantics based on resource types. For example a helm chart
can be download directly as helm chart archive, even if stored as OCI artifact.
This is handled by download handler. Their usage can be enabled with the <code>--download-handlers</code>
option. Otherwise the resource as returned by the access method is stored.


With the option <code>--recursive</code> the complete reference tree of a component reference is traversed.

\
If a component lookup for building a reference closure is required
the <code>--lookup</code>  option can be used to specify a fallback
lookup repository. By default, the component versions are searched in
the repository holding the component version for which the closure is
determined. For *Component Archives* this is never possible, because
it only contains a single component version. Therefore, in this scenario
this option must always be specified to be able to follow component
references.


If the verification store is enabled, resources downloaded from
signed or verified component versions are verified against their digests
provided by the component version.(not supported for using downloaders for the
resource download).

The usage of the verification store is enabled by <code>--check-verified</code> or by
specifying a verification file with <code>--verified</code>.

### SEE ALSO

#### Parents

* [ocm download](ocm_download.md)	 &mdash; Download oci artifacts, resources or complete components
* [ocm](ocm.md)	 &mdash; Open Component Model command line client



##### Additional Links

* [<b>ocm ocm-downloadhandlers</b>](ocm_ocm-downloadhandlers.md)	 &mdash; List of all available download handlers

