## plugin descriptor &mdash; Plugin Descriptor Format Description

### Description

The plugin descriptor describes the capabilities of a plugin. It uses the
following fields:

- **<code>version</code>** *string*

  The format version of the information descriptor. The actually supported
  version is <code>v1</code>

- **<code>pluginName</code>** *string*

  The name of the plugin, it must correspond to the file name of the executable.

- **<code>pluginVersion</code>** *string*

  The version of the plugin. This is just an information field not used by the
  library

- **<code>shortDescription</code>** *string*

  A short description shown in the plugin overview provided by the command
  <code>ocm get plugins</code>.

- **<code>description</code>** *string*

  A description explaining the capabilities of the plugin

- **<code>accessMethods</code>** *[]AccessMethodDescriptor*

  The list of access methods versions provided by this plugin.
  This feature is already used to establish new access types, if
  the plugins are registered at an OCM context.

- **<code>uploaders</code>** *[]UploaderDescriptor*

  The list of supported uploaders. Uploaders will be used in a future
  version to describe foreign repository targets for local blobs
  of dedicated types imported into an OCM registry.

- **<code>downloaders</code>** *[]DownloaderDescriptor*

  The list of supported downloaders. Downloaders will be used by the
  CLI download command to provide downloaded artifacts in a filesystem format
  applicable to the type specific tools, regardless of the format it is stored
  as blob in a component version. Therefore, they can be registered for
  combination of artifact type and optional mime type (describing the actually
  used blob format).

- **<code>actions</code>** *[]ActionDescriptor*

  The list of supported actions. Actions are defined by the used OCM
  library to externalize element or element type related tasks, which
  require dedicated environment specific actions.
  For example, the creation of OCI repositories before an artifact upload.

- **<code>valueMergeHandlers</code>** *[]ValueMergeHandlerDescriptor*

  The list of supported merge handlers. Merge handlers are used to
  merge label values if a component version is re-transferred to
  a target repository.

- **<code>labelMergeSpecifications</code>** *[]LabelMergeSpecification*

  The list of assignments of label merge specification to labels.

#### Access Method Descriptor

An access method descriptor describes a dedicated supported access method.
It uses the following fields:

- **<code>name</code>** *string*

  The name of the access method.

- **<code>version</code>** *string*

  The version of the access method (default is <code>v1</code>).

- **<code>description</code>** *string*

  The description of the dedicated kind of access method. It must
  only be reported for one supported version.

- **<code>format</code>** *string*

  The description of the dedicated format version of an access method.

- **<code>options</code>** *[]Option]*

  Optional list of options provided for the command <code>ocm add resources</code>.
  If options are given, the plugin must support the command [plugin accessmethod compose](plugin_accessmethod_compose.md).

  An option is defined by the following fields:

  - **<code>name</code>** *string*

    This required field describe the name of the option. THis might be
    the name of a preconfigured option, or a new one.

  - **<code>type</code>** *string*

    This optional field describe the intended type for a new option.

  - **<code>description</code>** *string*

    This optional field is as description for a newly created option.

  If possible, predefined standard options should be used. In such a case
  only the <code>name</code> field should be defined for
  an option. If required, new options can be defined by additionally specifying
  a type and a description. New options should be used very carefully. The
  chosen names MUST not conflict with names provided by other plugins. Therefore
  it is highly recommended to use names prefixed by the plugin name.


The following predefined option types can be used:


  - <code>accessComponent</code>: [*string*] component for access specification
  - <code>accessHostname</code>: [*string*] hostname used for access
  - <code>accessPackage</code>: [*string*] package or object name
  - <code>accessRegistry</code>: [*string*] registry base URL
  - <code>accessRepository</code>: [*string*] repository URL
  - <code>accessVersion</code>: [*string*] version for access specification
  - <code>artifactId</code>: [*string*] maven artifact id
  - <code>body</code>: [*string*] body of a http request
  - <code>bucket</code>: [*string*] bucket name
  - <code>classifier</code>: [*string*] maven classifier
  - <code>comment</code>: [*string*] comment field value
  - <code>commit</code>: [*string*] git commit id
  - <code>digest</code>: [*string*] blob digest
  - <code>extension</code>: [*string*] maven extension name
  - <code>globalAccess</code>: [*map[string]YAML*] access specification for global access
  - <code>groupId</code>: [*string*] maven group id
  - <code>header</code>: [*string:string,string*] http headers
  - <code>hint</code>: [*string*] (repository) hint for local artifacts
  - <code>identityPath</code>: [*[]identity*] identity path for specification
  - <code>idpath</code>: [*[]string*] identity path (attr=value{,attr=value}
  - <code>mediaType</code>: [*string*] media type for artifact blob representation
  - <code>noredirect</code>: [*bool*] http redirect behavior
  - <code>package</code>: [*string*] npm package name
  - <code>reference</code>: [*string*] reference name
  - <code>region</code>: [*string*] region name
  - <code>registry</code>: [*string*] npm package registry
  - <code>size</code>: [*int*] blob size
  - <code>url</code>: [*string*] artifact or server url
  - <code>verb</code>: [*string*] http request method
  - <code>version</code>: [*string*] npm package version

The following predefined value types are supported:


  - <code>YAML</code>: JSON or YAML document string
  - <code>[]byte</code>: byte value
  - <code>[]identity</code>: identity path
  - <code>[]string</code>: list of string values
  - <code>bool</code>: boolean flag
  - <code>int</code>: integer value
  - <code>map[string]YAML</code>: JSON or YAML map
  - <code>string</code>: string value
  - <code>string:string,string</code>: string map defined by dedicated assignment of comma separated strings
  - <code>string=YAML</code>: string map with arbitrary values defined by dedicated assignments
  - <code>string=string</code>: string map defined by dedicated assignments
  - <code>string=string,string</code>: string map defined by dedicated assignment of comma separated strings

#### Uploader Descriptor

The descriptor for an uploader has the following fields:

- **<code>name</code>** *string*

  The name of the uploader.

- **<code>description</code>** *string*

  The description of the uploader

- **<code>constraints</code>** *[]Constraint*

  The list of constraints the uploader is usable for. A constraint is described
  by two fields:

  - **<code>artifactType</code>** *string*

    Restrict the usage to a dedicated artifact type.

  - **<code>mediaType</code>** *string*

    Restrict the usage to a dedicated media type of the artifact blob.

  - **<code>contextType</code>** *string*

    Restrict the usage to a dedicated implementation backend technology.
    If specified, the attribute <code>repositoryType</code> must be set, also.

  - **<code>repositoryType</code>** *string*

    Restrict the usage to a dedicated implementation of the backend technology.
    If specified, the attribute <code>contextType</code> must be set, also.

#### Downloader Descriptor

The descriptor for a downloader has the following fields:

- **<code>name</code>** *string*

  The name of the uploader.

- **<code>description</code>** *string*

  The description of the uploader

- **<code>constraints</code>** *[]DownloadConstraint*

  The list of constraints the downloader is usable for. A constraint is described
  by two fields:

  - **<code>artifactType</code>** *string*

    Restrict the usage to a dedicated artifact type.

  - **<code>mediaType</code>** *string* (optional)

    Restrict the usage to a dedicated media type of the artifact blob.

#### Action Descriptor

The descriptor for an action has the following fields:

- **<code>name</code>** *string*

  The name of the action (for example <code>oci.repository.prepare</code>).

- **<code>versions</code>** *[]string*

  A list of accepted specification versions of the action.
  The used version is negotiated between the caller and the plugin
  by selecting the latest version supported by both parties.

- **<code>description</code>** *string* (optional)

  A short description of the provided tasks done by this action.

- **<code>defaultSelectors</code>** *[]string* (optional)

  A list of selectors, for which this action implementation is automatically
  be registered when the plugin is loaded. The selector syntax depends on
  the action type. (For example, the hostname (pattern) for the action
  <code>oci.repository.prepare</code>). The selectors are either directly matched
  with action requests or used as regular expression.

- **<code>consumerType</code>** *string* (optional)

  By default, the action gets access to the credentials provided for the
  element the action should work on. But it might be, that other credentials
  are required to fulfill its task. Therefore, the action can request a dedicated
  consumer type used to lookup the credentials. The consumer attributes are
  derived from the action specification and cannot be influenced by the
  plugin.

### Value Merge Handler Descriptor

The descriptor for a value merge handler has the following fields:

- **<code>name</code>** *string*

  The name of the algorithm.

- **<code>description</code>** *string*

  The description of the algorithm.

### Label Merge Specification

The descriptor for a label merge specification has the following fields:

- **<code>name</code>** *string*

  The name of the label.

- **<code>version</code>** *string* (optional)

  The dedicated label format version the specification should be used for. If no
  version is specified the setting is valid for all versions without a dedicated
  assignment.

- **<code>description</code>** *string*

  The details for the merging.

- **<code>algorithm</code>** *string*

  The name of (top-level) the algorithm to use.

- **<code>config</code>** *any* (optional)

  The configuration settings used for the algorithm. It may contain nested
  merge specifications.

### Examples

```json
{
  "version": "v1",
  "pluginName": "test",
  "pluginVersion": "v1",
  "shortDescription": "a test plugin",
  "description": "a test plugin with access method test",
  "accessMethods": [
    {
      "description": "",
      "name": "test",
      "shortDescription": "test access"
    },
    {
      "description": "",
      "name": "test",
      "shortDescription": "test access",
      "version": "v1"
    }
  ],
  "uploaders": [
    {
      "constraints": [
        {
          "artifactType": "TestArtifact"
        }
      ],
      "name": "testuploader"
    }
  ]
}
```

### SEE ALSO

#### Parents

* [plugin](plugin.md)	 &mdash; OCM Plugin



##### Additional Links

* [<b>plugin accessmethod compose</b>](plugin_accessmethod_compose.md)	 &mdash; compose access specification from options and base specification

