## plugin valuemergehandler execute &mdash; Execute A Value Merge

### Synopsis

```
plugin valuemergehandler execute <name> <spec> [<options>]
```

### Options

```
  -h, --help   help for execute
```

### Description


This command executes a a value merge. The values are taken from *stdin* as JSON
string. It has the following fields: 

- **<code>local</code>** *any*

  The local value to merge into the inbound value.

- **<code>inbound</code>** *any*

  The value to merge into. THis value is based on the original inbound value.

This action has to provide an execution result as JSON string on *stdout*. It has the 
following fields: 

- **<code>modified</code>** *bool*

  Whether the inbound value has been modified by merging with the local value.

- **<code>value</code>** *string*

  The merged value

- **<code>message</code>** *string*

  An error message.


### SEE ALSO

##### Parents

* [plugin valuemergehandler](plugin_valuemergehandler.md)	 &mdash; value merge handler operations
* [plugin](plugin.md)	 &mdash; OCM Plugin

