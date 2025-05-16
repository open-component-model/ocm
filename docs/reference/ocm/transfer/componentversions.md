---
title: "ocm transfer componentversions &mdash; Transfer Component Version"
url: "/docs/cli-reference/transfer/componentversions/"
sidebar:
  collapsed: true
---

### Synopsis

```bash
ocm transfer componentversions [<options>] {<component-reference>} <target>
```

#### Aliases

```text
componentversions, componentversion, cv, components, component, comps, comp, c
```

### Options

```text
  -B, --bom-file string             file name to write the component version BOM
  -c, --constraints constraints     version constraint
  -L, --copy-local-resources        transfer referenced local resources by-value
  -V, --copy-resources              transfer referenced resources by-value
      --copy-sources                transfer referenced sources by-value
      --disable-uploads             disable standard upload handlers for transport
      --enforce                     enforce transport as if target version were not present
  -h, --help                        help for componentversions
      --latest                      restrict component versions to latest
      --lookup stringArray          repository name or spec for closure lookup fallback
      --no-update                   don't touch existing versions in target
  -N, --omit-access-types strings   omit by-value transfer for resource types
  -f, --overwrite                   overwrite existing component versions
  -r, --recursive                   follow component reference nesting
      --repo string                 repository name or spec
      --script string               config name of transfer handler script
  -s, --scriptFile string           filename of transfer handler script
  -E, --stop-on-existing            stop on existing component version in target repository
  -t, --type string                 archive format (directory, tar, tgz) (default "directory")
      --uploader <name>=<value>     repository uploader (<name>[:<artifact type>[:<media type>[:<priority>]]]=<JSON target config>) (default [])
```

### Description

Transfer all component versions specified to the given target repository.
If only a component (instead of a component version) is specified all versions
are transferred.


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


The <code>--type</code> option accepts a file format for the
target archive to use. It is only evaluated if the target
archive does not exist yet. The following formats are supported:
- directory
- tar
- tgz

The default format is <code>directory</code>.


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


If the option <code>--overwrite</code> is given, component versions in the
target repository will be overwritten, if they already exist, but with different digest.
If the option <code>--enforce</code> is given, component versions in the
target repository will be transported as if they were not present on the target side,
regardless of their state (this is independent on their actual state, even identical
versions are re-transported).


With the option <code>--no-update</code> existing versions in the target
repository will not be touched at all. An additional specification of the
option <code>--overwrite</code> is ignored. By default, updates of
volatile (non-signature-relevant) information is enabled, but the
modification of non-volatile data is prohibited unless the overwrite
option is given.


If the option <code>--copy-resources</code> is given, all referential
resources will potentially be localized, mapped to component version local
resources in the target repository. If the option <code>--copy-local-resources</code>
is given, instead, only resources with the relation <code>local</code> will be
transferred. This behaviour can be further influenced by specifying a transfer
script with the <code>script</code> option family.


If the option <code>--copy-sources</code> is given, all referential
sources will potentially be localized, mapped to component version local
resources in the target repository.
This behaviour can be further influenced by specifying a transfer script
with the <code>script</code> option family.


If the option <code>--omit-access-types</code> is given, by-value transfer
is omitted completely for the given resource types.


If the option <code>--stop-on-existing</code> is given together with the <code>--recursive</code>
option, the recursion is stopped for component versions already existing in the
target repository. This behaviour can be further influenced by specifying a transfer script
with the <code>script</code> option family.



If the <code>--uploader</code> option is specified, appropriate uploader handlers
are configured for the operation. It has the following format

<center>
    <pre>&lt;name>:&lt;artifact type>:&lt;media type>=&lt;yaml target config></pre>
</center>

The uploader name may be a path expression with the following possibilities:
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


It is possible to use a dedicated transfer script based on spiff.
The option <code>--scriptFile</code> can be used to specify this script
by a file name. With <code>--script</code> it can be taken from the
CLI config using an entry of the following format:

<pre>
type: scripts.ocm.config.ocm.software
scripts:
  &lt;name>:
    path: &lt;filepath>
    script:
      &lt;scriptdata>
</pre>

Only one of the fields <code>path</code> or <code>script</code> can be used.

If no script option is given and the cli config defines a script <code>default</code>
this one is used.

### Examples

```bash
$ ocm transfer components -t tgz ghcr.io/open-component-model/ocm//ocm.software/ocmcli:0.17.0 ./ctf.tgz
$ ocm transfer components --latest -t tgz --repo OCIRegistry::ghcr.io/open-component-model/ocm ocm.software/ocmcli ./ctf.tgz
$ ocm transfer components --latest --copy-resources --type directory ghcr.io/open-component-model/ocm//ocm.software/ocmcli ./ctf
```

### SEE ALSO

#### Parents

* [ocm transfer](ocm_transfer.md)	 &mdash; Transfer artifacts or components
* [ocm](ocm.md)	 &mdash; Open Component Model command line client



##### Additional Links

* [<b>ocm ocm-uploadhandlers</b>](ocm_ocm-uploadhandlers.md)	 &mdash; List of all available upload handlers

