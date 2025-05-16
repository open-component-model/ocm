---
title: "ocm hash componentversions &mdash; Hash Component Version"
linkTitle: "hash componentversions"
url: "/docs/cli-reference/hash/componentversions/"
sidebar:
  collapsed: true
---

### Synopsis

```bash
ocm hash componentversions [<options>] {<component-reference>}
```

#### Aliases

```text
componentversions, componentversion, cv, components, component, comps, comp, c
```

### Options

```text
      --actual                    use actual component descriptor
  -c, --constraints constraints   version constraint
  -H, --hash string               hash algorithm (default "SHA-256")
  -h, --help                      help for componentversions
      --latest                    restrict component versions to latest
      --lookup stringArray        repository name or spec for closure lookup fallback
  -N, --normalization string      normalization algorithm (default "jsonNormalisation/v3")
  -O, --outfile string            Output file for normalized component descriptor (default "-")
  -o, --output string             output mode (JSON, json, norm, wide, yaml)
  -r, --recursive                 follow component reference nesting
      --repo string               repository name or spec
  -s, --sort stringArray          sort fields
  -U, --update                    update digests in component version
  -V, --verify                    verify digests found in component version
```

### Description

Hash lists normalized forms for all component versions specified, if only a component is specified
all versions are listed.


If the option <code>--constraints</code> is given, and no version is specified
for a component, only versions matching the given version constraints
(semver https://github.com/Masterminds/semver) are selected.
With <code>--latest</code> only
the latest matching versions will be selected.



If the option <code>--actual</code> is given the component descriptor actually
found is used as it is, otherwise the required digests are calculated on-the-fly.


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


The following normalization modes are supported with option <code>--normalization</code>:
  - <code>jsonNormalisation/v1</code>
  - <code>jsonNormalisation/v2</code>
  - <code>jsonNormalisation/v3</code> (default)


Note that the normalization algorithm is important to be equivalent when used for signing and verification, otherwise
the verification can fail. Please always migrate to the latest normalization algorithm whenever possible.
New signature algorithms can be used as soon as they are available in the component version after signing it.

The algorithms jsonNormalisation/v1 and jsonNormalisation/v2 are deprecated and should not be used anymore.
Please switch to jsonNormalisation/v3 as soon as possible.



The following hash modes are supported with option <code>--hash</code>:
  - <code>NO-DIGEST</code>
  - <code>SHA-256</code> (default)
  - <code>SHA-512</code>


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

With the option <code>--output</code> the output mode can be selected.
The following modes are supported:
  - <code></code> (default)
  - <code>JSON</code>
  - <code>json</code>
  - <code>norm</code>
  - <code>wide</code>
  - <code>yaml</code>

### Examples

```bash
$ ocm hash componentversion ghcr.io/open-component-model/ocm//ocm.software/ocmcli:0.17.0
$ ocm hash componentversion --repo OCIRegistry::ghcr.io/open-component-model/ocm ocm.software/ocmcli:0.17.0
```

### SEE ALSO

#### Parents

* [ocm hash](ocm_hash.md)	 &mdash; Hash and normalization operations
* [ocm](ocm.md)	 &mdash; Open Component Model command line client

