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
  applicable to the type specific tools, regatdless of the format it is stored
  as blob in a component version. Therefore they can be registered for
  combination of artifact type and optional mime type (describing the actually
  used blob format).

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
  it is highly recommended to use use names prefixed by the plugin name.


The following predefined option types can be used:


  - <code>accessHostname</code>: [*string*] hostname used for access
  - <code>accessRepository</code>: [*string*] repository URL
  - <code>accessVersion</code>: [*string*] version for access specification
  - <code>bucket</code>: [*string*] bucket name
  - <code>commit</code>: [*string*] git commit id
  - <code>digest</code>: [*string*] blob digest
  - <code>globalAccess</code>: [*map[string]YAML*] access specification for global access
  - <code>hint</code>: [*string*] (repository) hint for local artifacts
  - <code>mediaType</code>: [*string*] media type for artifact blob representation
  - <code>reference</code>: [*string*] reference name
  - <code>region</code>: [*string*] region name
  - <code>size</code>: [*int*] blob size

The following predefined value types are supported:


  - <code>YAML</code>: JSON or YAML document string
  - <code>[]string</code>: list of string values
  - <code>bool</code>: boolean flag
  - <code>int</code>: integer value
  - <code>map[string]YAML</code>: JSON or YAML map
  - <code>string</code>: string value
  - <code>string=YAML</code>: string map with arbitrary values defined by dedicated assignments
  - <code>string=string</code>: string map defined by dedicated assignments

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


### Examples

```
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

##### Parents

* [plugin](plugin.md)	 &mdash; OCM Plugin



##### Additional Links

* [<b>plugin accessmethod compose</b>](plugin_accessmethod_compose.md)	 &mdash; compose access specification from options and base specification

