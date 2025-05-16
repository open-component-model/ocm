---
title: "get componentversions"
menu:
  docs:
    parent: get
---
## ocm get componentversions &mdash; Get Component Version

### Synopsis

```bash
ocm get componentversions [<options>] {<component-reference>}
```

#### Aliases

```text
componentversions, componentversion, cv, components, component, comps, comp, c
```

### Options

```text
  -c, --constraints constraints   version constraint
  -h, --help                      help for componentversions
      --latest                    restrict component versions to latest
      --lookup stringArray        repository name or spec for closure lookup fallback
  -o, --output string             output mode (JSON, json, tree, wide, yaml)
  -r, --recursive                 follow component reference nesting
      --repo string               repository name or spec
  -S, --scheme string             schema version
  -s, --sort stringArray          sort fields
```

### Description

Get lists all component versions specified, if only a component is specified
all versions are listed.


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


If the option <code>--scheme</code> is given, the component descriptor
is converted to the specified format for output. If no format is given
the storage format of the actual descriptor is used or, for new ones v2
is used.
With <code>internal</code> the internal representation is shown.
The following schema versions are supported for explicit conversions:
  - <code>ocm.software/v3alpha1</code>
  - <code>v2</code>

With the option <code>--output</code> the output mode can be selected.
The following modes are supported:
  - <code></code> (default)
  - <code>JSON</code>
  - <code>json</code>
  - <code>tree</code>
  - <code>wide</code>
  - <code>yaml</code>

### Examples

```bash
$ ocm get componentversion ghcr.io/open-component-model/ocm//ocm.software/ocmcli:0.17.0
$ ocm get componentversion --repo OCIRegistry::ghcr.io/open-component-model/ocm ocm.software/ocmcli:0.17.0
```

### SEE ALSO

#### Parents

* [ocm get](ocm_get.md)	 &mdash; Get information about artifacts and components
* [ocm](ocm.md)	 &mdash; Open Component Model command line client

