## ocm add resources &mdash; Add Resources To A Component Version

### Synopsis

```
ocm add resources [<options>] <target> {<resourcefile> | <var>=<value>}
```

### Options

```
      --addenv                       access environment for templating
  -h, --help                         help for resources
  -s, --settings stringArray         settings file with variable settings (yaml)
      --templater string             templater to use (subst, spiff, go) (default "subst")
```


#### Access Specification Options

```
      --access YAML                  blob access specification (YAML)
      --accessHostname string        hostname used for access
      --accessRepository string      repository URL
      --accessType string            type of blob access specification
      --accessVersion string         version for access specification
      --bucket string                bucket name
      --commit string                git commit id
      --digest string                blob digest
      --globalAccess YAML            access specification for global access
      --hint string                  (repository) hint for local artifacts
      --mediaType string             media type for artifact blob representation
      --reference string             reference name
      --region string                region name
      --size int                     blob size
```


#### Input Specification Options

```
      --hint string                  (repository) hint for local artifacts
      --input YAML                   blob input specification (YAML)
      --inputCompress                compress option for input
      --inputExcludes stringArray    excludes (path) for inputs
      --inputFollowSymlinks          follow symbolic links during archive creation for inputs
      --inputIncludes stringArray    includes (path) for inputs
      --inputLibraries stringArray   library path for inputs
      --inputPath string             path field for input
      --inputPreserveDir             preserve directory in archive for inputs
      --inputType string             type of blob input specification
      --inputValues YAML             YAML based generic values for inputs
      --inputVariants stringArray    (platform) variants for inputs
      --inputVersion stringArray     version info for inputs
      --mediaType string             media type for artifact blob representation
```


#### Resource Meta Data Options

```
      --external                     flag non-local resource
      --extra <name>=<value>         resource extra identity (default [])
      --label <name>=<YAML>          resource label (leading * indicates signature relevant, optional version separated by @)
      --name string                  resource name
      --resource YAML                resource meta data (yaml)
      --type string                  resource type
      --version string               resource version
```

### Description


Add resources specified in a resource file to a component version.
So far only component archives are supported as target.

This command accepts  resource specification files describing the resources
to add to a component version. Elements must follow the resource meta data
description scheme of the component descriptor.

It is possible to describe a single resource via command line options.
The meta data of this element is described by the argument of option <code>--resource</code>,
which must be a YAML or JSON string.
Alternatively, the <em>name</em> and <em>version</em> can be specified with the
options <code>--name</code> and <code>--version</code>. With the option <code>--extra</code>
it is possible to add extra identity attributes. Explicitly specified options
override values specified by the <code>--resource</code> option.
(Note: Go templates are not supported for YAML-based option values. Besides
this restriction, the finally composed element description is still processd
by the selected templater.) 

The resource type can be specified with the option <code>--type</code>. Therefore, the
minimal required meta data for elements can be completely specified by dedicated
options and don't need the YAML option.

To describe the content of this element one of the options <code>--access</code> or
<code>--input</code> must be given. They take a YAML or JSON value describing an
attribute set, also. The structure of those values is similar to the <code>access</code>
or <code>input</code> fields of the description file format.
Non-local resources can be indicated using the option <code>--external</code>.

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

  The path must denote an image tag that can be found in the local
  docker daemon. The denoted image is packed as OCI artefact set.
  
  This blob type specification supports the following fields: 
  - **<code>path</code>** *string*
  
    This REQUIRED property describes the image name to import from the
    local docker daemon.
  
  - **<code>repository</code>** *string*
  
    This OPTIONAL property can be used to specify the repository hint for the
    generated local artefact access. It is prefixed by the component name if
    it does not start with slash "/".
  
  Options used to configure fields: <code>--hint</code>, <code>--inputPath</code>

- Input type <code>dockermulti</code>

  This input type describes the composition of a multi-platform OCI image.
  The various variants are taken from the local docker daemon. They should be 
  built with the buildx command for cross platform docker builds.
  The denoted images, as well as the wrapping image index is packed as OCI artefact set.
  
  This blob type specification supports the following fields:
  - **<code>variants</code>** *[]string*
  
    This REQUIRED property describes a set of  image names to import from the
    local docker daemon used to compose a resulting image index.
  
  - **<code>repository</code>** *string*
  
    This OPTIONAL property can be used to specify the repository hint for the
    generated local artefact access. It is prefixed by the component name if
    it does not start with slash "/".
  
  Options used to configure fields: <code>--hint</code>, <code>--inputVariants</code>

- Input type <code>file</code>

  The path must denote a file relative the resources file. 
  The content is compressed if the <code>compress</code> field
  is set to <code>true</code>.
  
  This blob type specification supports the following fields: 
  - **<code>path</code>** *string*
  
    This REQUIRED property describes the file path to the helm chart relative to the
    resource file location.
  
  - **<code>mediaType</code>** *string*
  
    This OPTIONAL property describes the media type to store with the local blob.
    The default media type is application/octet-stream and
    application/gzip if compression is enabled.
  
  - **<code>compress</code>** *bool*
  
    This OPTIONAL property describes whether the file content should be stored
    compressed or not.
  
  Options used to configure fields: <code>--inputCompress</code>, <code>--inputPath</code>, <code>--mediaType</code>

- Input type <code>helm</code>

  The path must denote an helm chart archive or directory
  relative to the resources file.
  The denoted chart is packed as an OCI artefact set.
  Additional provider info is taken from a file with the same name
  and the suffix <code>.prov</code>.
  
  If the chart should just be stored as archive, please use the 
  type <code>file</code> or <code>dir</code>.
  
  This blob type specification supports the following fields: 
  - **<code>path</code>** *string*
  
    This REQUIRED property describes the file path to the helm chart relative to the
    resource file location.
  
  - **<code>version</code>** *string*
  
    This OPTIONAL property can be set to configure an explicit version hint.
    If not specified the versio from the chart will be used.
    Basically, it is a good practice to use the component version for local resources
    This can be achieved by using templating for this attribute in the resource file.
  
  Options used to configure fields: <code>--inputCompress</code>, <code>--inputPath</code>, <code>--inputVersion</code>, <code>--mediaType</code>

- Input type <code>ociImage</code>

  The path must denote an OCI image reference.
  
  This blob type specification supports the following fields: 
  - **<code>path</code>** *string*
  
    This REQUIRED property describes the OVI image reference of the image to
    import.
  
  - **<code>repository</code>** *string*
  
    This OPTIONAL property can be used to specify the repository hint for the
    generated local artefact access. It is prefixed by the component name if
    it does not start with slash "/".
  
  Options used to configure fields: <code>--hint</code>, <code>--inputCompress</code>, <code>--inputPath</code>, <code>--mediaType</code>

- Input type <code>spiff</code>

  The path must denote a [spiff](https://github.com/mandelsoft/spiff) template relative the the resources file.
  The content is compressed if the <code>compress</code> field
  is set to <code>true</code>.
  
  This blob type specification supports the following fields: 
  - **<code>path</code>** *string*
  
    This REQUIRED property describes the file path to the helm chart relative to the
    resource file location.
  
  - **<code>mediaType</code>** *string*
  
    This OPTIONAL property describes the media type to store with the local blob.
    The default media type is application/octet-stream and
    application/gzip if compression is enabled.
  
  - **<code>compress</code>** *bool*
  
    This OPTIONAL property describes whether the file content should be stored
    compressed or not.
  
  - **<code>values</code>** *map[string]any*
  
    This OPTIONAL property describes an additional value binding for the template processing. It will be available
    under the node <code>inputvalues</code>.
  
  - **<code>libraries</code>** *[]string*
  
    This OPTIONAL property describes a list of spiff libraries to include in template
    processing.
  
  The variable settigs from the command line are available as binding, also. They are provided under the node
  <code>values</code>.
  
  Options used to configure fields: <code>--inputCompress</code>, <code>--inputLibraries</code>, <code>--inputPath</code>, <code>--inputValues</code>, <code>--mediaType</code>

The following access methods are known by the system.
Typically there is special support for the CLI artifact add commands.
The following types (with the field <code>type</code> in the <code>access</code> field
are handled:

- Access type <code>S3</code>

  This method implements the access of a blob stored in an S3 bucket.

  The following versions are supported:
  - Version <code>v1</code>
  
    The type specific specification fields are:
    
    - **<code>region</code>** (optional) *string*
    
      OCI repository reference (this artefact name used to store the blob).
    
    - **<code>bucket</code>** *string*
    
      The name of the S3 bucket containing the blob
    
    - **<code>key</code>** *string*
    
      The key of the desired blob
    
    Options used to configure fields: <code>--accessVersion</code>, <code>--bucket</code>, <code>--mediaType</code>, <code>--reference</code>, <code>--region</code>
  

- Access type <code>gitHub</code>

  This method implements the access of the content of a git commit stored in a
  GitHub repository.

  The following versions are supported:
  - Version <code>v1</code>
  
    The type specific specification fields are:
    
    - **<code>repoUrl</code>**  *string*
    
      Repository URL with or without scheme.
    
    - **<code>ref</code>** (optional) *string*
    
      Original ref used to get the commit from
    
    - **<code>commit</code>** *string*
    
      The sha/id of the git commit
    
    Options used to configure fields: <code>--accessHostname</code>, <code>--accessRepository</code>, <code>--commit</code>
  

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
    
      For example, if an OCI artefact originally referenced using the
      access method [<code>ociArtefact</code>](../../../../../docs/formats/accessmethods/ociArtefact.md) is stored during
      some transport step as local artefact, the reference name can be set
      to its original repository name. An import step into an OCI based OCM
      repository may then decide to make this artefact available again as 
      regular OCI artefact.
    
    - **<code>globalAccess</code>** (optional) *access method specification*
    
      If a resource blob is stored locally, the repository implementation
      may decide to provide an external access information (independent
      of the OCM model).
    
      For example, an OCI artefact stored as local blob
      can be additionally stored as regular OCI artefact in an OCI registry.
      
      This additional external access information can be added using
      a second external access method specification.
    
    Options used to configure fields: <code>--globalAccess</code>, <code>--hint</code>, <code>--mediaType</code>, <code>--reference</code>
  

- Access type <code>none</code>

  dummy resource with no access


- Access type <code>ociArtefact</code>

  This method implements the access of an OCI artefact stored in an OCI registry.

  The following versions are supported:
  - Version <code>v1</code>
  
    The type specific specification fields are:
    
    - **<code>imageReference</code>** *string*
    
      OCI image/artefact reference following the possible docker schemes:
      - <code>&lt;repo>/&lt;artefact>:&lt;digest>@&lt;tag></code>
      - <code><host>[&lt;port>]/&lt;repo path>/&lt;artefact>:&lt;version>@&lt;tag></code>
    
    Options used to configure fields: <code>--reference</code>
  

- Access type <code>ociBlob</code>

  This method implements the access of an OCI blob stored in an OCI repository.

  The following versions are supported:
  - Version <code>v1</code>
  
    The type specific specification fields are:
    
    - **<code>imageReference</code>** *string*
    
      OCI repository reference (this artefact name used to store the blob).
    
    - **<code>mediaType</code>** *string*
    
      The media type of the blob
    
    - **<code>digest</code>** *string*
    
      The digest of the blob used to access the blob in the OCI repository.
    
    - **<code>size</code>** *integer*
    
      The size of the blob
    
    Options used to configure fields: <code>--digest</code>, <code>--mediaType</code>, <code>--reference</code>, <code>--size</code>
  


### Examples



Add a resource directly by options

<pre>
$ ocm add resources path/to/ca &dash;&dash;name myresource &dash;&dash;type PlainText &dash;&dash;input '{ "type": "file", "path": "testdata/testcontent", "mediaType": "text/plain" }'
</pre>


Add a resource by a description file:

*resources.yaml*:

<pre>
&dash;&dash;&dash;
name: myrresource
type: PlainText
version: ${version]
input:
  type: file
  path: testdata/testcontent
  mediaType: text/plain
</pre>


<pre>
$ ocm add resources  path/to/ca  resources.yaml VERSION=1.0.0
</pre>



### SEE ALSO

##### Parents

* [ocm add](ocm_add.md)	 &mdash; Add resources or sources to a component archive
* [ocm](ocm.md)	 &mdash; Open Component Model command line client

