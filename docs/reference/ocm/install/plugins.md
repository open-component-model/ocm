---
title: "ocm install plugins &mdash; Install Or Update An OCM Plugin"
url: "/docs/cli-reference/install/plugins/"
sidebar:
  collapsed: true
---

### Synopsis

```bash
ocm install plugins [<options>] <component version ref> [<name>] | <name>
```

#### Aliases

```text
plugins, plugin, p
```

### Options

```text
  -c, --constraints constraints   version constraint
  -d, --describe                  describe plugin, only
  -f, --force                     overwrite existing plugin
  -h, --help                      help for plugins
  -r, --remove                    remove plugin
  -u, --update                    update plugin
```

### Description

Download and install a plugin provided by an OCM component version.
For the update mode only the plugin name is required.

If no version is specified the latest version is chosen. If at least one
version constraint is given, only the matching versions are considered.


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

### Examples

```bash
$ ocm install plugin ghcr.io/github.com/mandelsoft/cnudie//github.com/mandelsoft/ocmplugin:0.1.0-dev
$ ocm install plugin -c 1.2.x ghcr.io/github.com/mandelsoft/cnudie//github.com/mandelsoft/ocmplugin
$ ocm install plugin -u demo
$ ocm install plugin -r demo
```

### SEE ALSO

#### Parents

* [ocm install](ocm_install.md)	 &mdash; Install new OCM CLI components
* [ocm](ocm.md)	 &mdash; Open Component Model command line client

