---
title: "ocm get references &mdash; Get References Of A Component Version"
url: "/docs/cli-reference/get/references/"
sidebar:
  collapsed: true
---

### Synopsis

```bash
ocm get references [<options>]  <component> {<name> { <key>=<value> }}
```

#### Aliases

```text
references, reference, refs
```

### Options

```text
  -c, --constraints constraints   version constraint
  -h, --help                      help for references
      --latest                    restrict component versions to latest
      --lookup stringArray        repository name or spec for closure lookup fallback
  -o, --output string             output mode (JSON, json, tree, wide, yaml)
  -r, --recursive                 follow component reference nesting
      --repo string               repository name or spec
  -s, --sort stringArray          sort fields
```

### Description

Get references of a component version. References are specified
by identities. An identity consists of
a name argument followed by optional <code>&lt;key>=&lt;value></code>
arguments.


If the option <code>--constraints</code> is given, and no version is specified
for a component, only versions matching the given version constraints
(semver https://github.com/Masterminds/semver) are selected.
With <code>--latest</code> only
the latest matching versions will be selected.


If the <code>--repo</code> option is specified, the given names are interpreted
relative to the specified repository using the syntax

<center>
    <pre>&lt;component>[:&lt;version>]</pre>
</center>

If no <code>--repo</code> option is specified the given names are interpreted
as located OCM component version references:

<center>
    <pre>[&lt;repo type>::]&lt;host>[:&lt;port>][/&lt;base path>]//&lt;component>[:&lt;version>]</pre>
</center>

Additionally there is a variant to denote common transport archives
and general repository specifications

<center>
    <pre>[&lt;repo type>::]&lt;filepath>|&lt;spec json>[//&lt;component>[:&lt;version>]]</pre>
</center>

The <code>--repo</code> option takes an OCM repository specification:

<center>
    <pre>[&lt;repo type>::]&lt;configured name>|&lt;file path>|&lt;spec json></pre>
</center>

For the *Common Transport Format* the types <code>directory</code>,
<code>tar</code> or <code>tgz</code> is possible.

Using the JSON variant any repository types supported by the
linked library can be used:

OCI Repository types (using standard component repository to OCI mapping):

  - <code>CommonTransportFormat</code>: v1
  - <code>OCIRegistry</code>: v1
  - <code>oci</code>: v1
  - <code>ociRegistry</code>



With the option <code>--recursive</code> the complete reference tree of a component reference is traversed.

\
If a component lookup for building a reference closure is required
the <code>--lookup</code>  option can be used to specify a fallback
lookup repository. By default, the component versions are searched in
the repository holding the component version for which the closure is
determined. For *Component Archives* this is never possible, because
it only contains a single component version. Therefore, in this scenario
this option must always be specified to be able to follow component
references.

With the option <code>--output</code> the output mode can be selected.
The following modes are supported:
  - <code></code> (default)
  - <code>JSON</code>
  - <code>json</code>
  - <code>tree</code>
  - <code>wide</code>
  - <code>yaml</code>

### SEE ALSO

#### Parents

* [ocm get](ocm_get.md)	 &mdash; Get information about artifacts and components
* [ocm](ocm.md)	 &mdash; Open Component Model command line client

