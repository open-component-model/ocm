---
title: "ocm download cli - Download OCM CLI From An OCM Repository"
linkTitle: "download cli"
url: "/docs/cli-reference/download/cli/"
sidebar:
  collapsed: true
menu:
  docs:
    name: "download cli"
---

### Synopsis

```bash
ocm download cli [<options>]  [<component> {<name> { <key>=<value> }}]
```

#### Aliases

```text
cli, ocmcli, ocm-cli
```

### Options

```text
  -c, --constraints constraints   version constraint
  -h, --help                      help for cli
  -O, --outfile string            output file or directory
  -p, --path                      lookup executable in PATH
      --repo string               repository name or spec
      --use-verified              enable verification store
      --verified string           file used to remember verifications for downloads (default "~/.ocm/verified")
      --verify                    verify downloads
```

### Description

Download an OCM CLI executable. By default, the standard publishing component
and repository is used. Optionally, another component or repo and even a resource
can be specified. Resources are specified by identities. An identity consists of
a name argument followed by optional <code>&lt;key>=&lt;value></code>
arguments.

The option <code>-O</code> is used to declare the output destination.
The default location is the location of the <code>ocm</code> executable in
the actual PATH.


If the option <code>--constraints</code> is given, and no version is specified
for a component, only versions matching the given version constraints
(semver https://github.com/Masterminds/semver) are selected.


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



The library supports some downloads with semantics based on resource types. For example a helm chart
can be download directly as helm chart archive, even if stored as OCI artifact.
This is handled by download handler. Their usage can be enabled with the <code>--download-handlers</code>
option. Otherwise the resource as returned by the access method is stored.


If the verification store is enabled, resources downloaded from
signed or verified component versions are verified against their digests
provided by the component version.(not supported for using downloaders for the
resource download).

The usage of the verification store is enabled by <code>--use-verified</code> or by
specifying a verification file with <code>--verified</code>.

### SEE ALSO

#### Parents

* [ocm download](ocm_download.md)	 &mdash; Download oci artifacts, resources or complete components
* [ocm](ocm.md)	 &mdash; Open Component Model command line client

