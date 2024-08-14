## ocm execute action &mdash; Execute An Action

### Synopsis

```sh
ocm execute action [<options>] <action spec> {<cred>=<value>}
```

### Options

```
  -h, --help             help for action
  -m, --matcher string   matcher type override
  -n, --name string      action name (overrides type in specification)
  -o, --output string    output mode (json, yaml) (default "json")
```

### Description

Execute an action extension for a given action specification. The specification
show be a JSON or YAML argument.

Additional properties settings can be used to describe a consumer id
to retrieve credentials for.

### Examples

```
$ ocm execute action '{ "type": "oci.repository.prepare/v1", "hostname": "...", "repository": "..."}'
```

### SEE ALSO

#### Parents

* [ocm execute](ocm_execute.md)	 &mdash; Execute an element.
* [ocm](ocm.md)	 &mdash; Open Component Model command line client

