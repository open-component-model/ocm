---
title: "ocm ocm-accessmethods - List Of All Supported Access Methods"
linkTitle: "ocm-accessmethods"
url: "/docs/cli-reference/ocm-accessmethods/"
sidebar:
  collapsed: true
menu:
  docs:
    name: "ocm-accessmethods"
---

### Description

Access methods are used to handle the access to the content of artifacts
described in a component version. Therefore, an artifact entry contains
an access specification describing the access attributes for the dedicated
artifact.


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

### SEE ALSO

#### Parents

* [ocm](ocm.md)	 &mdash; Open Component Model command line client



##### Additional Links

* [<b>ocm get credentials</b>](ocm_get_credentials.md)	 &mdash; Get credentials for a dedicated consumer spec

