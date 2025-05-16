---
title: "ocm add sources &mdash; Add Source Information To A Component Version"
url: "/docs/cli-reference/add/sources/"
sidebar:
  collapsed: true
---

### Synopsis

```bash
ocm add sources [<options>] [<target>] {<resourcefile> | <var>=<value>}
```

#### Aliases

```text
sources, source, src, s
```

### Options

```text
      --addenv                              access environment for templating
      --dry-run                             evaluate and print source specifications
  -F, --file string                         target file/directory (default "component-archive")
  -h, --help                                help for sources
  -O, --output string                       output file for dry-run
  -P, --preserve-signature                  preserve existing signatures
  -R, --replace                             replace existing elements
  -s, --settings stringArray                settings file with variable settings (yaml)
      --templater string                    templater to use (go, none, spiff, subst) (default "subst")
```


#### Access Specification Options

```text
      --access YAML                         blob access specification (YAML)
      --accessComponent string              component for access specification
      --accessHostname string               hostname used for access
      --accessRepository string             repository or registry URL
      --accessType string                   type of blob access specification
      --accessVersion string                version for access specification
      --artifactId string                   maven artifact id
      --body string                         body of a http request
      --bucket string                       bucket name
      --classifier string                   maven classifier
      --commit string                       git commit id
      --digest string                       blob digest
      --extension string                    maven extension name
      --globalAccess YAML                   access specification for global access
      --groupId string                      maven group id
      --header <name>:<value>,<value>,...   http headers (default {})
      --hint string                         (repository) hint for local artifacts
      --identityPath {<name>=<value>}       identity path for specification
      --mediaType string                    media type for artifact blob representation
      --noredirect                          http redirect behavior
      --package string                      package or object name
      --reference string                    reference name
      --region string                       region name
      --size int                            blob size
      --url string                          artifact or server url
      --verb string                         http request method
```


#### Input Specification Options

```text
      --artifactId string                   maven artifact id
      --body string                         body of a http request
      --classifier string                   maven classifier
      --extension string                    maven extension name
      --groupId string                      maven group id
      --header <name>:<value>,<value>,...   http headers (default {})
      --hint string                         (repository) hint for local artifacts
      --identityPath {<name>=<value>}       identity path for specification
      --input YAML                          blob input specification (YAML)
      --inputComponent string               component name
      --inputCompress                       compress option for input
      --inputData !bytesBase64              data (string, !!string or !<base64>
      --inputExcludes stringArray           excludes (path) for inputs
      --inputFollowSymlinks                 follow symbolic links during archive creation for inputs
      --inputFormattedJson YAML             JSON formatted text
      --inputHelmRepository string          helm repository base URL
      --inputIncludes stringArray           includes (path) for inputs
      --inputJson YAML                      JSON formatted text
      --inputLibraries stringArray          library path for inputs
      --inputPath filepath                  path field for input
      --inputPlatforms stringArray          input filter for image platforms ([os]/[architecture])
      --inputPreserveDir                    preserve directory in archive for inputs
      --inputRepository string              repository or registry for inputs
      --inputText string                    utf8 text
      --inputType string                    type of blob input specification
      --inputValues YAML                    YAML based generic values for inputs
      --inputVariants stringArray           (platform) variants for inputs
      --inputVersion string                 version info for inputs
      --inputYaml YAML                      YAML formatted text
      --mediaType string                    media type for artifact blob representation
      --noredirect                          http redirect behavior
      --package string                      package or object name
      --url string                          artifact or server url
      --verb string                         http request method
```


#### Source Meta Data Options

```text
      --extra <name>=<value>                source extra identity (default [])
      --label <name>=<YAML>                 source label (leading * indicates signature relevant, optional version separated by @)
      --name string                         source name
      --source YAML                         source meta data (yaml)
      --type string                         source type
      --version string                      source version
```

### Description

Add information about the sources, e.g. commits in a Github repository,
that have been used to create the resources specified in a resource file to a component version.
So far only component archives are supported as target.

This command accepts source specification files describing the sources
to add to a component version. Elements must follow the source meta data
description scheme of the component descriptor. Besides referential sources
using the <code>access</code> attribute to describe the access method, it
is possible to describe local sources fed by local data using the <code>input</code>
field (see below).

The description file might contain:
- a single source
- a list of sources under the key <code>sources</code>
- a list of yaml documents with a single source or source list


It is possible to describe a single source via command line options.
The meta data of this element is described by the argument of option <code>--source</code>,
which must be a YAML or JSON string.
Alternatively, the <em>name</em> and <em>version</em> can be specified with the
options <code>--name</code> and <code>--version</code>. With the option <code>--extra</code>
it is possible to add extra identity attributes. Explicitly specified options
override values specified by the <code>--source</code> option.
(Note: Go templates are not supported for YAML-based option values. Besides
this restriction, the finally composed element description is still processed
by the selected template engine.)

The source type can be specified with the option <code>--type</code>. Therefore, the
minimal required meta data for elements can be completely specified by dedicated
options and don't need the YAML option.

To describe the content of this element one of the options <code>--access</code> or
<code>--input</code> must be given. They take a YAML or JSON value describing an
attribute set, also. The structure of those values is similar to the <code>access</code>
or <code>input</code> fields of the description file format.

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


The resource specification supports the following blob input types, specified
with the field <code>type</code> in the <code>input</code> field:

- Input type <code>binary</code>

  This blob type is used to provide base64 encoded binary content. The
  specification supports the following fields:
  - **<code>data</code>** *[]byte*

    The binary data to provide.

  - **<code>mediaType</code>** *string*

    This OPTIONAL property describes the media type to store with the local blob.
    The default media type is application/octet-stream and
    application/gzip if compression is enabled.

  - **<code>compress</code>** *bool*

    This OPTIONAL property describes whether the content should be stored
    compressed or not.

  Options used to configure fields: <code>--inputCompress</code>, <code>--inputData</code>, <code>--mediaType</code>

- Input type <code>dir</code>

  The path must denote a directory relative to the resources file, which is packed
  with tar and optionally compressed
  if the <code>compress</code> field is set to <code>true</code>. If the field
  <code>preserveDir</code> is set to true the directory itself is added to the tar.
  If the field <code>followSymLinks</code> is set to <code>true</code>, symbolic
  links are not packed but their targets files or folders.
  With the list fields <code>includeFiles</code> and <code>excludeFiles</code> it is
  possible to specify which files should be included or excluded. The values are
  regular expression used to match relative file paths. If no includes are specified
  all file not explicitly excluded are used.

  This blob type specification supports the following fields:
  - **<code>path</code>** *string*

    This REQUIRED property describes the file path to directory relative to the
    resource file location.

  - **<code>mediaType</code>** *string*

    This OPTIONAL property describes the media type to store with the local blob.
    The default media type is application/x-tar and
    application/gzip if compression is enabled.

  - **<code>compress</code>** *bool*

    This OPTIONAL property describes whether the file content should be stored
    compressed or not.

  - **<code>preserveDir</code>** *bool*

    This OPTIONAL property describes whether the specified directory with its
    basename should be included as top level folder.

  - **<code>followSymlinks</code>** *bool*

    This OPTIONAL property describes whether symbolic links should be followed or
    included as links.

  - **<code>excludeFiles</code>** *list of regex*

    This OPTIONAL property describes regular expressions used to match files
    that should NOT be included in the tar file. It takes precedence over
    the include match.

  - **<code>includeFiles</code>** *list of regex*

    This OPTIONAL property describes regular expressions used to match files
    that should be included in the tar file. If this option is not given
    all files not explicitly excluded are used.

  Options used to configure fields: <code>--inputCompress</code>, <code>--inputExcludes</code>, <code>--inputFollowSymlinks</code>, <code>--inputIncludes</code>, <code>--inputPath</code>, <code>--inputPreserveDir</code>, <code>--mediaType</code>

- Input type <code>docker</code>

  The path must denote an image tag that can be found in the local docker daemon.
  The denoted image is packed as OCI artifact set.
  The OCI image will contain an informational back link to the component version
  using the manifest annotation <code>software.ocm/component-version</code>.

  This blob type specification supports the following fields:
  - **<code>path</code>** *string*

    This REQUIRED property describes the image name to import from the
    local docker daemon.

  - **<code>repository</code>** *string*

    This OPTIONAL property can be used to specify the repository hint for the
    generated local artifact access. It is prefixed by the component name if
    it does not start with slash "/".

  Options used to configure fields: <code>--hint</code>, <code>--inputPath</code>

- Input type <code>dockermulti</code>

  This input type describes the composition of a multi-platform OCI image.
  The various variants are taken from the local docker daemon. They should be
  built with the "buildx" command for cross platform docker builds (see https://ocm.software/docs/tutorials/best-practices/#building-multi-architecture-images).
  The denoted images, as well as the wrapping image index, are packed as OCI
  artifact set.
  They will contain an informational back link to the component version
  using the manifest annotation <code>software.ocm/component-version</code>.

  This blob type specification supports the following fields:
  - **<code>variants</code>** *[]string*

    This REQUIRED property describes a set of  image names to import from the
    local docker daemon used to compose a resulting image index.

  - **<code>repository</code>** *string*

    This OPTIONAL property can be used to specify the repository hint for the
    generated local artifact access. It is prefixed by the component name if
    it does not start with slash "/".

  Options used to configure fields: <code>--hint</code>, <code>--inputVariants</code>

- Input type <code>file</code>

  The path must denote a file relative the resources file.
  The content is compressed if the <code>compress</code> field
  is set to <code>true</code>.

  This blob type specification supports the following fields:
  - **<code>path</code>** *string*

    This REQUIRED property describes the path to the file relative to the
    resource file location.

  - **<code>mediaType</code>** *string*

    This OPTIONAL property describes the media type to store with the local blob.
    The default media type is application/octet-stream and
    application/gzip if compression is enabled.

  - **<code>compress</code>** *bool*

    This OPTIONAL property describes whether the content should be stored
    compressed or not.

  Options used to configure fields: <code>--inputCompress</code>, <code>--inputPath</code>, <code>--mediaType</code>

- Input type <code>git</code>

  The repository type allows accessing an arbitrary git repository
  using the manifest annotation <code>software.ocm/component-version</code>.
  The ref can be used to further specify the branch or tag to checkout, otherwise the remote HEAD is used.

  This blob type specification supports the following fields:
  - **<code>repository</code>** *string*

    This REQUIRED property describes the URL of the git repository to access. All git URL formats are supported.

  - **<code>ref</code>** *string*

    This OPTIONAL property can be used to specify the remote branch or tag to checkout (commonly called ref).
    If not set, the default HEAD (remotes/origin/HEAD) of the remote is used.

  - **<code>commit</code>** *string*

    This OPTIONAL property can be used to specify the commit hash to checkout.
    If not set, the default HEAD of the ref is used.

  Options used to configure fields: <code>--inputRepository</code>, <code>--inputVersion</code>

- Input type <code>helm</code>

  The path must denote an helm chart archive or directory
  relative to the resources file or a chart name in a helm chart repository.
  The denoted chart is packed as an OCI artifact set.
  For the filesystem version additional provider info is taken from a file with
  the same name and the suffix <code>.prov</code>.

  If the chart should just be stored as plain archive, please use the
  type <code>file</code> or <code>dir</code>, instead.

  This blob type specification supports the following fields:
  - **<code>path</code>** *string*

    This REQUIRED property describes the file path to the helm chart relative to the
    resource file location.

  - **<code>version</code>** *string*

    This OPTIONAL property can be set to configure an explicit version hint.
    If not specified the version from the chart will be used.
    Basically, it is a good practice to use the component version for local resources
    This can be achieved by using templating for this attribute in the resource file.

  - **<code>helmRepository</code>** *string*

    This OPTIONAL property can be set, if the helm chart should be loaded from
    a helm repository instead of the local filesystem. It describes
    the base URL of the chart repository. If specified, the <code>path</code> field
    must describe the name of the chart in the chart repository, and <code>version</code>
    must describe the version of the chart imported from the chart repository

  - **<code>repository</code>** *string*

    This OPTIONAL property can be used to specify the repository hint for the
    generated local artifact access. It is prefixed by the component name if
    it does not start with slash "/".

  - **<code>caCertFile</code>** *string*

    This OPTIONAL property can be used to specify a relative filename for
    the TLS root certificate used to access a helm repository.

  - **<code>caCert</code>** *string*

    This OPTIONAL property can be used to specify a TLS root certificate used to
    access a helm repository.

  Options used to configure fields: <code>--hint</code>, <code>--inputCompress</code>, <code>--inputHelmRepository</code>, <code>--inputPath</code>, <code>--inputVersion</code>, <code>--mediaType</code>

- Input type <code>maven</code>

  The <code>repoUrl</code> is the url pointing either to the http endpoint of a maven
  repository (e.g. https://repo.maven.apache.org/maven2/) or to a file system based
  maven repository (e.g. file://local/directory).

  This blob type specification supports the following fields:
  - **<code>repoUrl</code>** *string*

    This REQUIRED property describes the url from which the resource is to be
    accessed.

  - **<code>groupId</code>** *string*

    This REQUIRED property describes the groupId of a maven artifact.

  - **<code>artifactId</code>** *string*
  	
    This REQUIRED property describes artifactId of a maven artifact.

  - **<code>version</code>** *string*

    This REQUIRED property describes the version of a maven artifact.

  - **<code>classifier</code>** *string*

    This OPTIONAL property describes the classifier of a maven artifact.

  - **<code>extension</code>** *string*

    This OPTIONAL property describes the extension of a maven artifact.

  Options used to configure fields: <code>--artifactId</code>, <code>--classifier</code>, <code>--extension</code>, <code>--groupId</code>, <code>--inputPath</code>, <code>--inputVersion</code>, <code>--url</code>

- Input type <code>npm</code>

  The <code>registry</code> is the url pointing to the npm registry from which a resource is
  downloaded.

  This blob type specification supports the following fields:
  - **<code>registry</code>** *string*

    This REQUIRED property describes the url from which the resource is to be
    downloaded.

  - **<code>package</code>** *string*
  	
    This REQUIRED property describes the name of the package to download.

  - **<code>version</code>** *string*

    This is an OPTIONAL property describing the version of the package to download. If
    not defined, latest will be used automatically.

  Options used to configure fields: <code>--inputRepository</code>, <code>--inputVersion</code>, <code>--package</code>

- Input type <code>ociArtifact</code>

  This input type is used to import an OCI image from an OCI registry.
  If it is a multi-arch image the set of platforms to be imported can be filtered using the "platforms"
  attribute. The path must denote an OCI image reference.

  This blob type specification supports the following fields:
  - **<code>path</code>** *string*

    This REQUIRED property describes the OCI image reference of the image to
    import.

  - **<code>repository</code>** *string*

    This OPTIONAL property can be used to specify the repository hint for the
    generated local artifact access. It is prefixed by the component name if
    it does not start with slash "/".

  - **<code>platforms</code>** *[]string*

    This OPTIONAL property can be used to filter index artifacts to include
    only images for dedicated operating systems/architectures.
    Elements must meet the syntax [&lt;os>]/[&lt;architecture>].

  Options used to configure fields: <code>--hint</code>, <code>--inputCompress</code>, <code>--inputPath</code>, <code>--inputPlatforms</code>, <code>--mediaType</code>

- Input type <code>ociImage</code>

  DEPRECATED: This type is deprecated, please use ociArtifact instead.

  Options used to configure fields: <code>--hint</code>, <code>--inputCompress</code>, <code>--inputPath</code>, <code>--inputPlatforms</code>, <code>--mediaType</code>

- Input type <code>ocm</code>

  This input type allows to get a resource artifact from an OCM repository.

  This blob type specification supports the following fields:
  - **<code>ocmRepository</code>** *repository specification*

    This REQUIRED property describes the OCM repository specification

  - **<code>component</code>** *string*

    This REQUIRED property describes the component na,e

  - **<code>version</code>** *string*

    This REQUIRED property describes the version of a maven artifact.

  - **<code>resourceRef</code>** *relative resource reference*

    This REQUIRED property describes the  resource reference for the desired
    resource relative to the given component version .

  Options used to configure fields: <code>--identityPath</code>, <code>--inputComponent</code>, <code>--inputRepository</code>, <code>--inputVersion</code>

- Input type <code>spiff</code>

  The path must denote a [spiff](https://github.com/mandelsoft/spiff) template relative the resources file.
  The content is compressed if the <code>compress</code> field
  is set to <code>true</code>.

  This blob type specification supports the following fields:
  - **<code>path</code>** *string*

    This REQUIRED property describes the path to the file relative to the
    resource file location.

  - **<code>mediaType</code>** *string*

    This OPTIONAL property describes the media type to store with the local blob.
    The default media type is application/octet-stream and
    application/gzip if compression is enabled.

  - **<code>compress</code>** *bool*

    This OPTIONAL property describes whether the content should be stored
    compressed or not.

  - **<code>values</code>** *map[string]any*

    This OPTIONAL property describes an additional value binding for the template processing. It will be available
    under the node <code>inputvalues</code>.

  - **<code>libraries</code>** *[]string*

    This OPTIONAL property describes a list of spiff libraries to include in template
    processing.

  The variable settings from the command line are available as binding, also. They are provided under the node
  <code>values</code>.

  Options used to configure fields: <code>--inputCompress</code>, <code>--inputLibraries</code>, <code>--inputPath</code>, <code>--inputValues</code>, <code>--mediaType</code>

- Input type <code>utf8</code>

  This blob type is used to provide inline text based content (UTF8). The
  specification supports the following fields:
  - **<code>text</code>** *string*

    The utf8 string content to provide.

  - **<code>json</code>** *JSON or JSON string interpreted as JSON*

    The content emitted as JSON.

  - **<code>formattedJson</code>** *YAML/JSON or JSON/YAML string interpreted as JSON*

    The content emitted as formatted JSON.

  - **<code>yaml</code>** *AML/JSON or JSON/YAML string interpreted as YAML*

    The content emitted as YAML.

  - **<code>mediaType</code>** *string*

    This OPTIONAL property describes the media type to store with the local blob.
    The default media type is application/octet-stream and
    application/gzip if compression is enabled.

  - **<code>compress</code>** *bool*

    This OPTIONAL property describes whether the content should be stored
    compressed or not.

  Options used to configure fields: <code>--inputCompress</code>, <code>--inputFormattedJson</code>, <code>--inputJson</code>, <code>--inputText</code>, <code>--inputYaml</code>, <code>--mediaType</code>

- Input type <code>wget</code>

  The <code>url</code> is the url pointing to the http endpoint from which a resource is
  downloaded. The <code>mimeType</code> can be used to specify the MIME type of the
  resource.

  This blob type specification supports the following fields:
  - **<code>url</code>** *string*

    This REQUIRED property describes the url from which the resource is to be
    downloaded.

  - **<code>mediaType</code>** *string*

    This OPTIONAL property describes the media type of the resource to be
    downloaded. If omitted, ocm tries to read the mediaType from the Content-Type header
    of the http response. If the mediaType cannot be set from the Content-Type header as well,
    ocm tries to deduct the mediaType from the URL. If that is not possible either, the default
    media type is defaulted to application/octet-stream.

  - **<code>header</code>** *map[string][]string*
  	
    This OPTIONAL property describes the http headers to be set in the http request to the server.

  - **<code>verb</code>** *string*

    This OPTIONAL property describes the http verb (also known as http request method) for the http
    request. If omitted, the http verb is defaulted to GET.

  - **<code>body</code>** *[]byte*

    This OPTIONAL property describes the http body to be included in the request.

  - **<code>noredirect</code>** *bool*

    This OPTIONAL property describes whether http redirects should be disabled. If omitted,
    it is defaulted to false (so, per default, redirects are enabled).

  Options used to configure fields: <code>--body</code>, <code>--header</code>, <code>--mediaType</code>, <code>--noredirect</code>, <code>--url</code>, <code>--verb</code>

The following list describes the supported access methods, their versions
and specification formats.
Typically there is special support for the CLI artifact add commands.
The access method specification can be put below the <code>access</code> field.
If always requires the field <code>type</code> describing the kind and version
shown below.

- Access type <code>git</code>

  This method implements the access of the content of a git commit stored in a
  Git repository.

  The following versions are supported:
  - Version <code>v1alpha1</code>

    The type specific specification fields are:

    - **<code>repoUrl</code>**  *string*

      Repository URL with or without scheme.

    - **<code>ref</code>** (optional) *string*

      Original ref used to get the commit from

    - **<code>commit</code>** *string*

      The sha/id of the git commit

  Options used to configure fields: <code>--accessRepository</code>, <code>--commit</code>, <code>--reference</code>

- Access type <code>gitHub</code>

  This method implements the access of the content of a git commit stored in a
  GitHub repository.

  The following versions are supported:
  - Version <code>v1</code>

    The type specific specification fields are:

    - **<code>repoUrl</code>**  *string*

      Repository URL with or without scheme.

    - **<code>ref</code>** (optional) *string*

      Original ref used to get the commit from. Mutually exclusive with <code>ref</code>.

    - **<code>commit</code>** *string*

      The sha/id of the git commit. Mutually exclusive with <code>commit</code>.

  Options used to configure fields: <code>--accessHostname</code>, <code>--accessRepository</code>, <code>--commit</code>, <code>--reference</code>

- Access type <code>helm</code>

  This method implements the access of a Helm chart stored in a Helm repository.

  The following versions are supported:
  - Version <code>v1</code>

    The type specific specification fields are:

    - **<code>helmRepository</code>** *string*

      Helm repository URL.

    - **<code>helmChart</code>** *string*

      The name of the Helm chart and its version separated by a colon.

    - **<code>version</code>** *string*

      The version of the Helm chart if not specified as part of the chart name.

    - **<code>caCert</code>** *string*

      An optional TLS root certificate.

    - **<code>keyring</code>** *string*

      An optional keyring used to verify the chart.

    It uses the consumer identity type HelmChartRepository with the fields
    for a hostpath identity matcher (see [ocm get credentials](ocm_get_credentials.md)).

  Options used to configure fields: <code>--accessRepository</code>, <code>--accessVersion</code>, <code>--package</code>

- Access type <code>localBlob</code>

  This method is used to store a resource blob along with the component descriptor
  on behalf of the hosting OCM repository.

  Its implementation is specific to the implementation of OCM
  repository used to read the component descriptor. Every repository
  implementation may decide how and where local blobs are stored,
  but it MUST provide an implementation for this method.

  Regardless of the chosen implementation the attribute specification is
  defined globally the same.

  The following versions are supported:
  - Version <code>v1</code>

    The type specific specification fields are:

    - **<code>localReference</code>** *string*

      Repository type specific location information as string. The value
      may encode any deep structure, but typically just an access path is sufficient.

    - **<code>mediaType</code>** *string*

      The media type of the blob used to store the resource. It may add
      format information like <code>+tar</code> or <code>+gzip</code>.

    - **<code>referenceName</code>** (optional) *string*

      This optional attribute may contain identity information used by
      other repositories to restore some global access with an identity
      related to the original source.

      For example, if an OCI artifact originally referenced using the
      access method <code>ociArtifact</code> is stored during
      some transport step as local artifact, the reference name can be set
      to its original repository name. An import step into an OCI based OCM
      repository may then decide to make this artifact available again as
      regular OCI artifact.

    - **<code>globalAccess</code>** (optional) *access method specification*

      If a resource blob is stored locally, the repository implementation
      may decide to provide an external access information (independent
      of the OCM model).

      For example, an OCI artifact stored as local blob
      can be additionally stored as regular OCI artifact in an OCI registry.

      This additional external access information can be added using
      a second external access method specification.

  Options used to configure fields: <code>--globalAccess</code>, <code>--hint</code>, <code>--mediaType</code>, <code>--reference</code>

- Access type <code>maven</code>

  This method implements the access of a Maven artifact in a Maven repository.

  The following versions are supported:
  - Version <code>v1</code>

    The type specific specification fields are:

    - **<code>repoUrl</code>** *string*

      URL of the Maven repository

    - **<code>groupId</code>** *string*

      The groupId of the Maven artifact

    - **<code>artifactId</code>** *string*

      The artifactId of the Maven artifact

    - **<code>version</code>** *string*

      The version name of the Maven artifact

    - **<code>classifier</code>** *string*

      The optional classifier of the Maven artifact

    - **<code>extension</code>** *string*

      The optional extension of the Maven artifact

  Options used to configure fields: <code>--accessRepository</code>, <code>--accessVersion</code>, <code>--artifactId</code>, <code>--classifier</code>, <code>--extension</code>, <code>--groupId</code>

- Access type <code>none</code>

  dummy resource with no access


- Access type <code>npm</code>

  This method implements the access of an NPM package in an NPM registry.

  The following versions are supported:
  - Version <code>v1</code>

    The type specific specification fields are:

    - **<code>registry</code>** *string*

      Base URL of the NPM registry.

    - **<code>package</code>** *string*

      The name of the NPM package

    - **<code>version</code>** *string*

      The version name of the NPM package

  Options used to configure fields: <code>--accessRepository</code>, <code>--accessVersion</code>, <code>--package</code>

- Access type <code>ociArtifact</code>

  This method implements the access of an OCI artifact stored in an OCI registry.

  The following versions are supported:
  - Version <code>v1</code>

    The type specific specification fields are:

    - **<code>imageReference</code>** *string*

      OCI image/artifact reference following the possible docker schemes:
      - <code>&lt;repo>/&lt;artifact>:&lt;digest>@&lt;tag></code>
      - <code><host>[&lt;port>]/&lt;repo path>/&lt;artifact>:&lt;version>@&lt;tag></code>

  Options used to configure fields: <code>--reference</code>

- Access type <code>ociBlob</code>

  This method implements the access of an OCI blob stored in an OCI repository.

  The following versions are supported:
  - Version <code>v1</code>

    The type specific specification fields are:

    - **<code>imageReference</code>** *string*

      OCI repository reference (this artifact name used to store the blob).

    - **<code>mediaType</code>** *string*

      The media type of the blob

    - **<code>digest</code>** *string*

      The digest of the blob used to access the blob in the OCI repository.

    - **<code>size</code>** *integer*

      The size of the blob

  Options used to configure fields: <code>--digest</code>, <code>--mediaType</code>, <code>--reference</code>, <code>--size</code>

- Access type <code>ocm</code>

  This method implements the access of any resource artifact stored in an OCM
  repository. Only repository types supporting remote access should be used.

  The following versions are supported:
  - Version <code>v1</code>

    The type specific specification fields are:

    - **<code>ocmRepository</code>** *json*

      The repository spec for the OCM repository

    - **<code>component</code>** *string*

      *(Optional)* The name of the component. The default is the
      own component.

    - **<code>version</code>** *string*

      *(Optional)* The version of the component. The default is the
      own component version.

    - **<code>resourceRef</code>** *relative resource ref*

      The resource reference of the denoted resource relative to the
      given component version.

    It uses the consumer identity and credentials for the intermediate repositories
    and the final resource access.

  Options used to configure fields: <code>--accessComponent</code>, <code>--accessRepository</code>, <code>--accessVersion</code>, <code>--identityPath</code>

- Access type <code>s3</code>

  This method implements the access of a blob stored in an S3 bucket.

  The following versions are supported:
  - Version <code>v1</code>

    The type specific specification fields are:

    - **<code>region</code>** (optional) *string*

      OCI repository reference (this artifact name used to store the blob).

    - **<code>bucket</code>** *string*

      The name of the S3 bucket containing the blob

    - **<code>key</code>** *string*

      The key of the desired blob

    - **<code>version</code>** (optional) *string*

      The key of the desired blob

    - **<code>mediaType</code>** (optional) *string*

      The media type of the content

  - Version <code>v2</code>

    The type specific specification fields are:

    - **<code>region</code>** (optional) *string*

      OCI repository reference (this artifact name used to store the blob).

    - **<code>bucketName</code>** *string*

      The name of the S3 bucket containing the blob

    - **<code>objectKey</code>** *string*

      The key of the desired blob

    - **<code>version</code>** (optional) *string*

      The key of the desired blob

    - **<code>mediaType</code>** (optional) *string*

      The media type of the content

  Options used to configure fields: <code>--accessVersion</code>, <code>--bucket</code>, <code>--mediaType</code>, <code>--reference</code>, <code>--region</code>

- Access type <code>wget</code>

  This method implements access to resources stored on an http server.

  The following versions are supported:
  - Version <code>v1</code>

    The <code>url</code> is the url pointing to the http endpoint from which a resource is
    downloaded. The <code>mimeType</code> can be used to specify the MIME type of the
    resource.

    This blob type specification supports the following fields:
    - **<code>url</code>** *string*

    This REQUIRED property describes the url from which the resource is to be
    downloaded.

    - **<code>mediaType</code>** *string*

    This OPTIONAL property describes the media type of the resource to be
    downloaded. If omitted, ocm tries to read the mediaType from the Content-Type header
    of the http response. If the mediaType cannot be set from the Content-Type header as well,
    ocm tries to deduct the mediaType from the URL. If that is not possible either, the default
    media type is defaulted to application/octet-stream.

    - **<code>header</code>** *map[string][]string*

    This OPTIONAL property describes the http headers to be set in the http request to the server.

    - **<code>verb</code>** *string*

    This OPTIONAL property describes the http verb (also known as http request method) for the http
    request. If omitted, the http verb is defaulted to GET.

    - **<code>body</code>** *[]byte*

    This OPTIONAL property describes the http body to be included in the request.

    - **<code>noredirect</code>** *bool*

    This OPTIONAL property describes whether http redirects should be disabled. If omitted,
    it is defaulted to false (so, per default, redirects are enabled).

  Options used to configure fields: <code>--body</code>, <code>--header</code>, <code>--mediaType</code>, <code>--noredirect</code>, <code>--url</code>, <code>--verb</code>



The <code>--replace</code> option allows users to specify whether adding an
element with the same name and extra identity but different version as an
existing element, append (false) or replace (true) the existing element.

The <code>--preserve-signature</code> option prohibits changes of signature
relevant elements.


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


### Examples

```bash
$ ocm add sources --file path/to/cafile sources.yaml
```

### SEE ALSO

#### Parents

* [ocm add](ocm_add.md)	 &mdash; Add elements to a component repository or component version
* [ocm](ocm.md)	 &mdash; Open Component Model command line client



##### Additional Links

* [<b>ocm get credentials</b>](ocm_get_credentials.md)	 &mdash; Get credentials for a dedicated consumer spec

