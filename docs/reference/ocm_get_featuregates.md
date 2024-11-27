## ocm get featuregates &mdash; List Feature Gates

### Synopsis

```bash
ocm get featuregates [<options>] {<name>}
```

#### Aliases

```text
featuregates, fg
```

### Options

```text
  -h, --help               help for featuregates
  -o, --output string      output mode (JSON, json, wide, yaml)
  -s, --sort stringArray   sort fields
```

### Description

Show feature gates and the activation.

The following feature gates are supported:


With the option <code>--output</code> the output mode can be selected.
The following modes are supported:
  - <code></code> (default)
  - <code>JSON</code>
  - <code>json</code>
  - <code>wide</code>
  - <code>yaml</code>

### Examples

```bash
$ ocm get featuregates
```

### SEE ALSO

#### Parents

* [ocm get](ocm_get.md)	 &mdash; Get information about artifacts and components
* [ocm](ocm.md)	 &mdash; Open Component Model command line client

