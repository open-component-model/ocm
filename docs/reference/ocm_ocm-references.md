---
title: "ocm references"
menu:
  docs:
    parent: cli-reference
---
## ocm ocm-references &mdash; Notation For OCM References

### Description

The command line client supports a special notation scheme for specifying
references to OCM components and repositories. This allows for specifying
references to any registry supported by the OCM toolset that can host OCM
components:

<center>
    <pre>[+][&lt;type>::][./]&lt;file path>//&lt;component id>[:&lt;version>]</pre>
</center>

or

<center>
	<pre>[+][&lt;type>::][&lt;json repo spec>//]&lt;component id>[:&lt;version>]</pre>
</center>

or

<center>
    <pre>[+][&lt;type>::][&lt;scheme>://]&lt;domain>[:&lt;port>][/&lt;repository prefix>]//&lt;component id>[:&lt;version]</pre>
</center>

or

<center>
	<pre>[+][&lt;type>::][&lt;scheme>://]&lt;host>[:&lt;port>][/&lt;repository prefix>]//&lt;component id>[:&lt;version]</pre>
</center>

---

Besides dedicated components it is also possible to denote repositories
as a whole:

<center>
    <pre>[+][&lt;type>::][./]&lt;file path></pre>
</center>

or

<center>
    <pre>[+][&lt;type>::]&lt;json repo spec></pre>
</center>

or

<center>
    <pre>[+][&lt;type>::][&lt;scheme>://]&lt;domain>[:&lt;port>][/&lt;repository prefix>]</pre>
</center>

or

<center>
	<pre>[+][&lt;type>::][&lt;scheme>://]&lt;host>[:&lt;port>][/&lt;repository prefix>]</pre>
</center>

The optional <code>+</code> is used for file based implementations
(Common Transport Format) to indicate the creation of a not yet existing
file.

The **type** may contain a file format qualifier separated by a <code>+</code>
character. The following formats are supported: <code>directory</code>, <code>tar</code>, <code>tgz</code>
### Examples

```text
Complete Component Reference Specifications (including all optional arguments):

+ctf+directory::./ocm/ctf//ocm.software/ocmcli:0.7.0

oci::{"baseUrl":"ghcr.io","componentNameMapping":"urlPath","subPath":"open-component-model"}//ocm.software/ocmcli.0.7.0

oci::https://ghcr.io:443/open-component-model//ocm.software/ocmcli:0.7.0

oci::http://localhost:8080/local-component-repository//ocm.software/ocmcli:0.7.0

---

Short-Hand Component Reference Specifications (omitting optional arguments):

./ocm/ctf//ocm.software/ocmcli:0.7.0

ghcr.io/open-component-model//ocm.software/ocmcli:0.7.0

localhost:8080/local-component-repository//ocm.software/ocmcli:0.7.0 (defaulting to https)

http://localhost:8080/local-component-repository//ocm.software/ocmcli:0.7.0
```

### SEE ALSO

#### Parents

* [ocm](ocm.md)	 &mdash; Open Component Model command line client

