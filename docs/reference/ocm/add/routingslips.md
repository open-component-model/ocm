---
title: "ocm add routingslips &mdash; Add Routing Slip Entry"
url: "/docs/cli-reference/add/routingslips/"
sidebar:
  collapsed: true
---

### Synopsis

```bash
ocm add routingslips [<options>] <component-version> <routing-slip> <type>
```

#### Aliases

```text
routingslips, routingslip, rs
```

### Options

```text
  -S, --algorithm string     signature handler (default "RSASSA-PKCS1-V1_5")
      --digest string        parent digest to use
  -h, --help                 help for routingslips
      --links strings        links to other slip/entries (<slipname>[@<digest>])
      --lookup stringArray   repository name or spec for closure lookup fallback
      --repo string          repository name or spec
```


#### Entry Specification Options

```text
      --comment string       comment field value
      --entry YAML           routing slip entry specification (YAML)
```

### Description

Add a routing slip entry for the specified routing slip name to the given
component version. The name is typically a DNS domain name followed by some
qualifiers separated by a slash (/). It is possible to use arbitrary types,
the type is not checked, if it is not known. Accordingly, an arbitrary config
given as JSON or YAML can be given to determine the attribute set of the new
entry for unknown types.


The following list describes the well-known entry types explicitly supported
by this version of the CLI, their versions and specification formats. Other
kinds of entries can be configured using the <code>--entry</code> option.

- Entry type <code>comment</code>

  An unstructured comment as entry in a routing slip.

  The following versions are supported:
  - Version <code>v1</code>

    The type specific specification fields are:

    - **<code>comment</code>**  *string*

      Any text as entry in a routing slip.

  Options used to configure fields: <code>--comment</code>


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
$ ocm add routingslip ghcr.io/mandelsoft/ocm//ocmdemoinstaller:0.0.1-dev mandelsoft.org comment --entry "comment=some text"
```

### SEE ALSO

#### Parents

* [ocm add](ocm_add.md)	 &mdash; Add elements to a component repository or component version
* [ocm](ocm.md)	 &mdash; Open Component Model command line client

