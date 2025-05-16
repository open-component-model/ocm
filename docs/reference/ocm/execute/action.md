---
title: "execute action"
url: "/docs/cli-reference/execute/action/"
---

## ocm execute action &mdash; Execute An Action

### Synopsis

```bash
ocm execute action [<options>] <action spec> {<cred>=<value>}
```

### Options

```text
  -h, --help             help for action
  -m, --matcher string   matcher type override
  -n, --name string      action name (overrides type in specification)
  -o, --output string    output mode (json, yaml) (default "json")
```

### Description

Execute an action extension for a given action specification. The specification
should be a JSON or YAML argument.

Additional properties settings can be used to describe a consumer id
to retrieve credentials for.

The following actions are supported:
- Name: oci.repository.prepare
    Prepare the usage of a repository in an OCI registry.

    The hostname of the target repository is used as selector. The action should
    assure, that the requested repository is available on the target OCI registry.

    Spec version v1 uses the following specification fields:
    - <code>hostname</code> *string*: The  hostname of the OCI registry.
    - <code>repository</code> *string*: The OCI repository name.

  Possible Consumer Attributes:
  - <code>hostname</code>
  - <code>port</code>
  - <code>pathprefix</code>

### Examples

```bash
$ ocm execute action '{ "type": "oci.repository.prepare/v1", "hostname": "...", "repository": "..."}'
```

### SEE ALSO

#### Parents

* [ocm execute](ocm_execute.md)	 &mdash; Execute an element.
* [ocm](ocm.md)	 &mdash; Open Component Model command line client

