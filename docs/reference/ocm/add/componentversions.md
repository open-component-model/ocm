---
title: "ocm add componentversions - Add Component Version(S) To A (New) Transport Archive"
linkTitle: "add componentversions"
url: "/docs/cli-reference/add/componentversions/"
sidebar:
  collapsed: true
menu:
  docs:
    name: "add componentversions"
---

### Synopsis

```bash
ocm add componentversions [<options>] [--version <version>] [<ctf archive>] {<component-constructor.yaml>}
```

#### Aliases

```text
componentversions, componentversion, cv, components, component, comps, comp, c
```

### Options

```text
      --addenv                    access environment for templating
  -C, --complete                  include all referenced component version
  -L, --copy-local-resources      transfer referenced local resources by-value
  -V, --copy-resources            transfer referenced resources by-value
  -c, --create                    (re)create archive
      --dry-run                   evaluate and print component specifications
  -F, --file string               target file/directory (default "transport-archive")
  -f, --force                     remove existing content
  -h, --help                      help for componentversions
      --lookup stringArray        repository name or spec for closure lookup fallback
  -O, --output string             output file for dry-run
  -P, --preserve-signature        preserve existing signatures
  -R, --replace                   replace existing elements
  -S, --scheme string             schema version (default "v2")
  -s, --settings stringArray      settings file with variable settings (yaml)
      --skip-digest-generation    skip digest creation
      --templater string          templater to use (go, none, spiff, subst) (default "subst")
  -t, --type string               archive format (directory, tar, tgz) (default "directory")
      --uploader <name>=<value>   repository uploader (<name>[:<artifact type>[:<media type>[:<priority>]]]=<JSON target config>) (default [])
  -v, --version string            default version for components
```

### Description

Add component versions specified by a constructor file to a Common Transport
Archive. The archive might be either a directory prepared to host component version
content or a tar/tgz file (see option --type).

If option <code>--create</code> is given, the archive is created first. An
additional option <code>--force</code> will recreate an empty archive if it
already exists.

If option <code>--complete</code> is given all component versions referenced by
the added one, will be added, also. Therefore, the <code>--lookup</code> is required
to specify an OCM repository to lookup the missing component versions. If
additionally the <code>-V</code> is given, the resources of those additional
components will be added by value.


The <code>--replace</code> option allows users to specify whether adding an
element with the same name and extra identity but different version as an
existing element, append (false) or replace (true) the existing element.

The <code>--preserve-signature</code> option prohibits changes of signature
relevant elements.


The source, resource and reference list can be composed according to the commands
[ocm add sources](ocm_add_sources.md), [ocm add resources](ocm_add_resources.md), [ocm add references](ocm_add_references.md),
respectively.

The description file might contain:
- a single component as shown in the example
- a list of components under the key <code>components</code>
- a list of yaml documents with a single component or component list

The optional field <code>meta.configuredSchemaVersion</code> for a component
entry can be used to specify a dedicated serialization format to use for the
component descriptor. If given it overrides the <code>--schema</code> option
of the command. By default, v2 is used.

Various elements support to add arbitrary information by using labels
(see [ocm ocm-labels](ocm_ocm-labels.md)).


The <code>--type</code> option accepts a file format for the
target archive to use. It is only evaluated if the target
archive does not exist yet. The following formats are supported:
- directory
- tar
- tgz

The default format is <code>directory</code>.


All yaml/json defined resources can be templated.
Variables are specified as regular arguments following the syntax <code>&lt;name>=&lt;value></code>.
Additionally settings can be specified by a yaml file using the <code>--settings <file></code>
option. With the option <code>--addenv</code> environment variables are added to the binding.
Values are overwritten in the order environment, settings file, command line settings.

Note: Variable names are case-sensitive.

Example:
<pre>
&lt;command> &lt;options> -- MY_VAL=test &lt;args>
</pre>

There are several templaters that can be selected by the <code>--templater</code> option:
- <code>go</code> go templating supports complex values.

  <pre>
    key:
      subkey: "abc {{.MY_VAL}}"
  </pre>

- <code>none</code> do not do any substitution.

- <code>spiff</code> [spiff templating](https://github.com/mandelsoft/spiff).

  It supports complex values. the settings are accessible using the binding <code>values</code>.
  <pre>
    key:
      subkey: "abc (( values.MY_VAL ))"
  </pre>

- <code>subst</code> simple value substitution with the <code>drone/envsubst</code> templater.

  It supports string values, only. Complex settings will be json encoded.
  <pre>
    key:
      subkey: "abc ${MY_VAL}"
  </pre>


\
If a component lookup for building a reference closure is required
the <code>--lookup</code>  option can be used to specify a fallback
lookup repository. By default, the component versions are searched in
the repository holding the component version for which the closure is
determined. For *Component Archives* this is never possible, because
it only contains a single component version. Therefore, in this scenario
this option must always be specified to be able to follow component
references.


If the option <code>--copy-resources</code> is given, all referential
resources will potentially be localized, mapped to component version local
resources in the target repository. If the option <code>--copy-local-resources</code>
is given, instead, only resources with the relation <code>local</code> will be
transferred. This behaviour can be further influenced by specifying a transfer
script with the <code>script</code> option family.



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

### Examples




<pre>
$ ocm add componentversions &dash;&dash;file ctf &dash;&dash;version 1.0 component&dash;constructor.yaml
</pre>


and a file <code>component-constructor.yaml</code>:


<pre>
name: ocm.software/demo/test
version: 1.0.0
provider:
  name: ocm.software
  labels:
    &dash; name: city
      value: Karlsruhe
labels:
  &dash; name: purpose
    value: test

resources:
  &dash; name: text
    type: PlainText
    input:
      type: file
      path: testdata
  &dash; name: data
    type: PlainText
    input:
      type: binary
      data: IXN0cmluZ2RhdGE=

</pre>


The resource <code>text</code> is taken from a file <code>testdata</code> located
next to the description file.


### SEE ALSO

#### Parents

* [ocm add](ocm_add.md)	 &mdash; Add elements to a component repository or component version
* [ocm](ocm.md)	 &mdash; Open Component Model command line client



##### Additional Links

* [<b>ocm add sources</b>](ocm_add_sources.md)	 &mdash; add source information to a component version
* [<b>ocm add resources</b>](ocm_add_resources.md)	 &mdash; add resources to a component version
* [<b>ocm add references</b>](ocm_add_references.md)	 &mdash; add aggregation information to a component version
* [<b>ocm ocm-labels</b>](ocm_ocm-labels.md)	 &mdash; Labels and Label Merging
* [<b>ocm ocm-uploadhandlers</b>](ocm_ocm-uploadhandlers.md)	 &mdash; List of all available upload handlers

