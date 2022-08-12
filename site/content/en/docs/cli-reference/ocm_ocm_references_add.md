
---
title: ocm_ocm_references_add
url: /docs/cli-reference/ocm_ocm_references_add/
date: 2022-08-12T11:14:49+01:00
draft: false
images: []
menu:
  docs:
    parent: cli-reference
toc: true
---
## ocm ocm references add &mdash; Add Aggregation Information To A Component Version

### Synopsis

```
ocm ocm references add [<options>] <target> {<resourcefile> | <var>=<value>}
```

### Options

```
      --addenv                 access environment for templating
  -h, --help                   help for add
  -s, --settings stringArray   settings file with variable settings (yaml)
      --templater string       templater to use (subst, spiff, go) (default "subst")
```

### Description


Add  aggregation information specified in a resource file to a component version.
So far only component archives are supported as target.

Templating:
All yaml/json defined resources can be templated.
Variables are specified as regular arguments following the syntax <code>&lt;name>=&lt;value></code>.
Additionally settings can be specified by a yaml file using the <code>--settings <file></code>
option. With the option <code>--addenv</code> environment variables are added to the binding.
Values are overwritten in the order environment, settings file, commmand line settings. 

Note: Variable names are case-sensitive.

Example:
<pre>
<command> <options> -- MY_VAL=test <args>
</pre>

There are several templaters that can be selected by the <code>--templater</code> option:
- envsubst: simple value substitution with the <code>drone/envsubst</code> templater. It
  supports string values, only. Complext settings will be json encoded.
  <pre>
  key:
    subkey: "abc ${MY_VAL}"
  </pre>

- go: go templating supports complex values.
  <pre>
  key:
    subkey: "abc {{.MY_VAL}}"
  </pre>

- spiff: [spiff templating](https://github.com/mandelsoft/spiff) supports
  complex values. the settings are accessible using the binding <tt>values</tt>.
  <pre>
  key:
    subkey: "abc (( values.MY_VAL ))"
  </pre>

This command accepts reference specification files describing the references
to add to a component version.

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
  regular expression used to match relative file paths. If no inlcudes are specified
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

- Input type <code>docker</code>
  
  The path must denote an image tag that can be found in the local
  docker daemon. The denoted image is packed an OCI artefact set.
  
  This blob type specification supports the following fields: 
  - **<code>path</code>** *string*
  
    This REQUIRED property describes the image name to import from the
    local docker daemon.

- Input type <code>file</code>
  
  The path must denote a file relative the the resources file.
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



### SEE ALSO

##### Parents

* [ocm ocm references](ocm_ocm_references.md)	 - Commands related to component references in component versions
* [ocm ocm](ocm_ocm.md)	 - Dedicated command flavors for the Open Component Model
* [ocm](ocm.md)	 - Open Component Model command line client

