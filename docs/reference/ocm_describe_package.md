---
title: "describe package"
menu:
  docs:
    parent: describe
---
## ocm describe package &mdash; Describe TOI Package

### Synopsis

```bash
ocm describe package [<options>] {<component-reference>} {<resource id field>}
```

#### Aliases

```text
package, pkg, componentversion, cv, component, comp, c
```

### Options

```text
  -h, --help                 help for package
      --lookup stringArray   repository name or spec for closure lookup fallback
      --repo string          repository name or spec
```

### Description

Describe a TOI package provided by a resource of an OCM component version.

The package resource must have the type <code>toiPackage</code>.
This is a simple YAML file resource describing the bootstrapping of a dedicated kind
of software. See also the topic [ocm toi-bootstrapping](ocm_toi-bootstrapping.md).

The first matching resource of this type is selected. Optionally a set of
identity attribute can be specified used to refine the match. This can be the
resource name and/or other key/value pairs (<code>&lt;attr>=&lt;value></code>).


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

\
If a component lookup for building a reference closure is required
the <code>--lookup</code>  option can be used to specify a fallback
lookup repository. By default, the component versions are searched in
the repository holding the component version for which the closure is
determined. For *Component Archives* this is never possible, because
it only contains a single component version. Therefore, in this scenario
this option must always be specified to be able to follow component
references.

### Examples

```bash
$ ocm toi describe package ghcr.io/mandelsoft/ocm//ocmdemoinstaller:0.0.1-dev
```

### SEE ALSO

#### Parents

* [ocm describe](ocm_describe.md)	 &mdash; Describe various elements by using appropriate sub commands.
* [ocm](ocm.md)	 &mdash; Open Component Model command line client



##### Additional Links

* [<b>ocm toi-bootstrapping</b>](ocm_toi-bootstrapping.md)	 &mdash; Tiny OCM Installer based on component versions

