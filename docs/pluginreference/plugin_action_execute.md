## plugin action execute &mdash; Execute An Action

### Synopsis

```
plugin action execute <spec> [<options>]
```

### Options

```
  -C, --credential <name>=<value>   dedicated credential value (default [])
  -c, --credentials YAML            credentials
  -h, --help                        help for execute
```

### Description


This command executes an action.

This action has to provide an execution result as JSON string on *stdout*. It has the 
following fields: 

- **<code>name</code>** *string*

  The name and version of the action result. It must match the value
  from the action specification.

- **<code>message</code>** *string*

  An error message.

Additional fields depend on the kind of action.


### SEE ALSO

##### Parents

* [plugin action](plugin_action.md)	 &mdash; action operations
* [plugin](plugin.md)	 &mdash; OCM Plugin

