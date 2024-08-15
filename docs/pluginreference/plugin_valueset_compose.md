## plugin valueset compose &mdash; Compose Value Set From Options And Base Specification

### Synopsis

```bash
plugin valueset compose <purpose> <name> <options json> <base spec json> [<options>]
```

### Options

```
  -h, --help   help for compose
```

### Description

The task of this command is ued to compose and validate a value set based on
some explicitly given input options and preconfigured specifications.

The finally composed set has to be returned as JSON document
on *stdout*.

This command is only used, if for a value set descriptor configuration
na direct composition rules are configured ([plugin descriptor](plugin_descriptor.md)).

If possible, predefined standard options should be used. In such a case only the
<code>name</code> field should be defined for an option. If required, new options can be
defined by additionally specifying a type and a description. New options should
be used very carefully. The chosen names MUST not conflict with names provided
by other plugins. Therefore, it is highly recommended to use names prefixed
by the plugin name.


The following predefined option types can be used:


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
  - <code>mediaType</code>: [*string*] media type for artifact blob representation
  - <code>noredirect</code>: [*bool*] http redirect behavior
  - <code>reference</code>: [*string*] reference name
  - <code>region</code>: [*string*] region name
  - <code>size</code>: [*int*] blob size
  - <code>url</code>: [*string*] artifact or server url
  - <code>verb</code>: [*string*] http request method

The following predefined value types are supported:


  - <code>YAML</code>: JSON or YAML document string
  - <code>[]byte</code>: byte value
  - <code>[]string</code>: list of string values
  - <code>bool</code>: boolean flag
  - <code>int</code>: integer value
  - <code>map[string]YAML</code>: JSON or YAML map
  - <code>string</code>: string value
  - <code>string:string,string</code>: string map defined by dedicated assignment of comma separated strings
  - <code>string=YAML</code>: string map with arbitrary values defined by dedicated assignments
  - <code>string=string</code>: string map defined by dedicated assignments
  - <code>string=string,string</code>: string map defined by dedicated assignment of comma separated strings
### SEE ALSO

#### Parents

* [plugin valueset](plugin_valueset.md)	 &mdash; valueset operations
* [plugin](plugin.md)	 &mdash; OCM Plugin



##### Additional Links

* [<b>plugin descriptor</b>](plugin_descriptor.md)	 &mdash; Plugin Descriptor Format Description
