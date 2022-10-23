## ocm add resource-configuration &mdash; Add A Resource Specification To A Resource Config File

### Synopsis

```
ocm add resource-configuration [<options>] <target> {<configfile> | <var>=<value>}
```

### Options

```
      --access YAML                  blob access specification (YAML)
      --accessHostname string        hostname used for access
      --accessRepository string      repository URL
      --accessType string            type of blob access specification
      --accessVersion string         version for access specification
      --addenv                       access environment for templating
      --bucket string                bucket name
      --commit string                git commit id
      --digest string                blob digest
      --external                     flag non-local resource
      --extra <name>=<value>         resource extra identity (default [])
      --globalAccess YAML            access specification for global access
  -h, --help                         help for resource-configuration
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
      --label <name>=<YAML>          resource label (leading * indicates signature relevant, optional version separated by @)
      --mediaType string             media type for artifact blob representation
      --name string                  resource name
      --reference string             reference name
      --region string                region name
      --resource YAML                resource meta data (yaml)
  -s, --settings stringArray         settings file with variable settings (yaml)
      --size int                     blob size
      --templater string             templater to use (subst, spiff, go) (default "none")
      --type string                  resource type
      --version string               resource version
```

### Description


Add a resource specification to a resource config file used by [ocm add resources](ocm_add_resources.md).

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
Non-local resources can be indicated using the option <code>--external</code>. Elements must follow the resource meta data
description scheme of the component descriptor.

If expressions/templates are used in the specification file an appropriate
templater and the required settings might be required to provide
a correct input validation.

This command accepts additional resource specification files describing the sources
to add to a component version.


Templating:
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
  
  Options used to configure fields: <code>--inputIncludes</code>, <code>--inputExcludes</code>, <code>--inputPreserveDir</code>, <code>--inputFollowSymlinks</code>, <code>--inputPath</code>, <code>--mediaType</code>, <code>--inputCompress</code>


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
  
  Options used to configure fields: <code>--inputVariants</code>, <code>--hint</code>


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
  
  Options used to configure fields: <code>--inputPath</code>, <code>--mediaType</code>, <code>--inputCompress</code>


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
  
  Options used to configure fields: <code>--inputPath</code>, <code>--inputVersion</code>, <code>--mediaType</code>, <code>--inputCompress</code>


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
  
  Options used to configure fields: <code>--inputPath</code>, <code>--hint</code>, <code>--mediaType</code>, <code>--inputCompress</code>


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
  
  Options used to configure fields: <code>--inputLibraries</code>, <code>--inputValues</code>, <code>--inputPath</code>, <code>--mediaType</code>, <code>--inputCompress</code>



### Examples

```
$ ocm add resource-config resources.yaml --name myresource --type PlainText --input '{ "type": "file", "path": "testdata/testcontent", "mediaType": "text/plain" }'
```

### SEE ALSO

##### Parents

* [ocm add](ocm_add.md)	 &mdash; Add resources or sources to a component archive
* [ocm](ocm.md)	 &mdash; Open Component Model command line client

