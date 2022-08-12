
---
title: ocm_oci_oci-references
url: /docs/cli-reference/ocm_oci_oci-references/
date: 2022-08-12T11:14:49+01:00
draft: false
images: []
menu:
  docs:
    parent: cli-reference
toc: true
---
## ocm oci oci-references &mdash; Notation For OCI References

### Description


The command line client supports a special notation scheme for specifying
references to instances of oci like registries. This allows for specifying
references to any registry supported by the OCM toolset that can host OCI
artefacts. As a subset the regular OCI artefact notation used for docker
images are possible:

<center>
    <pre>[+][&lt;type>::][./][&lt;file path>//&lt;repository>[:&lt;tag>][@&lt;digest>]</pre>
        or
    <pre>[&lt;type>::][&lt;json repo spec>//]&lt;repository>[:&lt;tag>][@&lt;digest>]</pre>
        or
    <pre>[&lt;type>::][&lt;scheme>:://]&lt;domain>[:&lt;port>/]&lt;repository>[:&lt;tag>][@&lt;digest>]</pre>
        or
    <pre>&lt;docker library>[:&lt;tag>][@&lt;digest>]</pre>
        or
    <pre>&lt;docker repository>/&lt;docker image>[:&lt;tag>][@&lt;digest>]</pre>
</center>

Besides dedicated artefacts it is also possible to denote registries
as a whole:

<center>
    <pre>[+][&lt;type>::][&lt;scheme>:://]&lt;domain>[:&lt;port>]</pre>
        or
    <pre>[+][&lt;type>::]&lt;json repo spec></pre>
        or
    <pre>[+][&lt;type>::][./]&lt;file path></pre>
</center>

The optional <code>+</code> is used for file based implementations
(Common Transport Format) to indicate the creation of a not yet existing
file.

The **type** may contain a file format qualifier separated by a <code>+</code>
character. The following formats are supported: <code>directory</code>, <code>tar</code>, <code>tgz</code>

### Examples

```

ghcr.io/mandelsoft/cnudie:1.0.0

```

### SEE ALSO

##### Parents

* [ocm oci](ocm_oci.md)	 - Dedicated command flavors for the OCI layer
* [ocm](ocm.md)	 - Open Component Model command line client

