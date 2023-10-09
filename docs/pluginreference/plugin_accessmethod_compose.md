## plugin accessmethod compose &mdash; Compose Access Specification From Options And Base Specification

### Synopsis

```
plugin accessmethod compose <name> <options json> <base spec json> [<options>]
```

### Options

```
  -h, --help   help for compose
```

### Description


The task of this command is to compose an access specification based on some
explicitly given input options and preconfigured specifications.

The finally composed access specification has to be returned as JSON document
on *stdout*.

This command is only used, if for an access method descriptor configuration
options are defined ([plugin descriptor](plugin_descriptor.md)).

If possible, predefined standard options should be used. In such a case only the
<code>name</code> field should be defined for an option. If required, new options can be
defined by additionally specifying a type and a description. New options should
be used very carefully. The chosen names MUST not conflict with names provided
by other plugins. Therefore, it is highly recommended to use use names prefixed
by the plugin name.


The following predefined option types can be used:


  - <code>accessHostname</code>: [*string*] hostname used for access
  - <code>accessPackage</code>: [*string*] package or object name
  - <code>accessRegistry</code>: [*string*] registry base URL
  - <code>accessRepository</code>: [*string*] repository URL
  - <code>accessVersion</code>: [*string*] version for access specification
  - <code>bucket</code>: [*string*] bucket name
  - <code>comment</code>: [*string*] comment field value
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

### SEE ALSO

##### Parents

* [plugin accessmethod](plugin_accessmethod.md)	 &mdash; access method operations
* [plugin](plugin.md)	 &mdash; OCM Plugin



##### Additional Links

* [<b>plugin descriptor</b>](plugin_descriptor.md)	 &mdash; Plugin Descriptor Format Description

