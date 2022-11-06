# Extending the Library by Plugins

The library has several extension points,which can be used by a registration
mechanism to add further variants, like repository types, backend technologies,
access methods, blob downloaders and uploaders.

This requires Go coding, which is feasible for additional standard
implementations. Nevertheless, it is useful to provide a more dynamic 
way to enrich the functionality of the library and the OCM command line
tool.

This can be achieved by the experimental *plugin* concept. It allows 
to implement functionality in separate executables (the plugins) and
register them for any main program based on this library.

## Commands

A plugin must provide a set of commands to implement the intended extension.

The library allows to configure a configuration for a plugin, this configuration
is optionally passed to all commands as JSON argument using option `-c`.

Errors have to be reported on *stderr* as JSON string with the fields:

- **`error`** *string*

  The error message provided by a command.


### `info` (Plugin Info)

**Synopsis:** `<plugin> [-c <pluginconfig>] info`

The capabilities provided by a plugin are queried using the
command `info`.

It must respond with JSON *Plugin Descriptor* on standard output 

#### Plugin Descriptor

The plugin descriptor describes the capabilities of a plugin. It uses the
following fields:

- **`version`** *string*

  The format version of the information descriptor. The actually supported
  version is `v1`

- **`pluginName`** *string*

  The name of the plugin, it must correspond to the file name of the executable.

- **`pluginVersion`** *string*

  The version of the plugin. This is just an information field not used by the 
  library

- **`shortDescription`** *string*

  A short description shown in the plugin overview provided by the command 
  `ocm ger plugins`.

- **`description`** *string*

  A description explaining the capabilities of the plugin

- **`accessMethods`** *[]AccessMethodDescriptor*

  The list of access methods versions provided by this plugin.
  This feature is already used to establish new access types, if
  the plugins are registered at an OCM context.

- **`uploaders`** *[]UploaderDescriptor*
  
  The list of supported uploaders. Uploaders will be used in a future
  version to describe foreign repository targets for local blobs
  of dedicated types imported into an OCM registry.

- **`downloaders`** *[]DownloaderDescriptor*

  The list of supported downloaders. Downloaders will be used by the
  CLI download command to provide downloaded artifacts in a filesystem format
  applicable to the type specific tools, regatdless of the format it is stored
  as blob in a component version. Therefore they can be registered for
  combination of artifact type and optional mime type (describing the actually
  used blob format).

#### Access Method Descriptor

An access method descriptor describes a dedicated supported access method.
It uses the following fields:

- **`name`** *string*

  The name of the access method.

- **`version`** *string*

  The version of the access method (default is `v1`).

- **`description`** *string*

  The description of the dedicated kind of access method. It must
  only be reported for one supported version.

- **`format`** *string*

  The description of the dedicated format version of an access method.

#### Uploader Descriptor

The descriptor for an uploader has the following fields:

- **`name`** *string*

  The name of the uploader.

- **`description`** *string*

  The description of the uploader

- **`constraints`** *[]Constraint*

  The list of constraints the uploader is usable for. A constraint is described 
  by two fields:
  
  - **`artifactType`** *string*
    
    Restrict the usage to a dedicated artifact type.

  - **`mediaType`** *string*
  
    Restrict the usage to a dedicated media type of the artifact blob.

  - **`contextType`** *string*

    Restrict the usage to a dedicated implementation backend technology.
    If specified, the attribute `repositoryType` must be set, also.

  - **`repositoryType`** *string*

    Restrict the usage to a dedicated implementation of the backend technology.
    If specified, the attribute `contextType` must be set, also.

#### Downloader Descriptor

The descriptor for a downloader has the following fields:

- **`name`** *string*

  The name of the uploader.

- **`description`** *string*

  The description of the uploader

- **`constraints`** *[]DownloadConstraint*

  The list of constraints the downloader is usable for. A constraint is described
  by two fields:

  - **`artifactType`** *string*

    Restrict the usage to a dedicated artifact type.

  - **`mediaType`** *string* (optional)

    Restrict the usage to a dedicated media type of the artifact blob.

### `accessmethods` (Access Method related Commands)

This command group provides all commands used to implement an access method
described by an access method descriptor. It requires the following 
nested commands:

#### `validate` (Validate an Access Specification)

**Synopsis:** `<plugin> [-c <pluginconfig>] accessmethod validate <spec>`

This command accepts an access specification as argument. It is used to
validate the specification and to provide some metadata for the given
specification.

This metadata has to be provided as JSON string on *stdout* and has the 
following fields: 

- **`mediaType`** *string*

  The media type of the artifact described by the specification. It may be part
  of the specification or implicitly determined by the access method.

- **`description`** *string*

  A short textual description of the described location.

- **`hint`** *string*

  A name hint of the described location used to reconstruct a useful
  name for local blobs uploaded to a dedicated repository technology.

- **`consumerId`** *map[string]string*

  The consumer id used to determine optional credentials for the 
  underlying repository. If specified, at least the `type` field must be set.


#### `get` (Get the Blob described by an Access Specification)

**Synopsis:** `<plugin> [-c <pluginconfig>] accessmethod get <options> <spec>`

**Options**:

```
  -C, --credential <name>=<value>   dedicated credential value (default [])
  -c, --credentials YAML            credentials
```

Return the blob described by the given access method on *stdout* 


### `upload` (Uploder related Commands))

This command group provides all commands used to implement an uploader
described by an uploader descriptor. It requires the following
nested commands:

#### `validate` (Validate an Upload Target Specification)

**Synopsis:** `<plugin> [-c <pluginconfig>] upload validate <name> <spec>`

**Options:**


```
  -a, --artifactType string   artifact type of input blob
  -m, --mediaType string      media type of input blob
```

This command accepts a target specification as argument. It is used to
validate the specification and to provide some metadata for the given
specification.

This metadata has to be provided as JSON string on *stdout* and has the
following fields:

- **`consumerId`** *map[string]string*

  The consumer id used to determine optional credentials for the
  underlying repository. If specified, at least the `type` field must be set.

#### `put` (Store a Blob in the described Target)

**Synopsis:** `<plugin> [-c <pluginconfig>] upload put <options> <name> <spec>`

**Options**:

```
  -a, --artifactType string         artifact type of input blob
  -C, --credential <name>=<value>   dedicated credential value (default [])
  -c, --credentials YAML            credentials
  -H, --hint string                 reference hint for storing blob
  -m, --mediaType string            media type of input blob
```

Read the blob content from *stdin*, store the blob and return the
access specification (as JSON string) usable to retrieve the blob, again,
on * stdout* 

### `download` (Download an Artifact Bob to a (set of) filesystem file(s))

**Synopsis:** `<plugin> [-c <pluginconfig>] download <name> <targetpath>`

**Options:**

```
  -a, --artifactType string   artifact type of input blob
  -m, --mediaType string      media type of input blob
```

This command accepts a target filepath as argument. It is used as base name
to store the downloaded content. The blob content is provided on the
standard input.

The task of this command is to transform the content of the provided 
blob into a filesystem structure applicable to the tools working
with content of the given artifact type.


## Implementation support

This library provides a command frame in package `pkg/contexts/ocm/plugin/ppi`.
It implements all the required command based on some interfaces, which must be
implemented by a plugin. These implementations are registered at a 
*Plugin*, which can then be passed to the standard implementation.

An example can be found in [`cmds/demoplugin`](https://github.com/open-component-model/ocm/tree/main/cmds/demoplugin).