## ocm ocm ocm-references &mdash; Notation For OCM References

### Description


The command line client supports a special notation scheme for specifying
references to OCM components and repositories. This allows for specifying
references to any registry supported by the OCM toolset that can host OCM
components:

<center>
    <pre>[+][&lt;type>::][./][&lt;file path>//&lt;component id>[:&lt;version>]</pre>
        or
    <pre>[+][&lt;type>::]&lt;domain>[:&lt;port>][/&lt;repository prefix>]//&lt;component id>[:&lt;version]</pre>
        or
    <pre>[&lt;type>::][&lt;json repo spec>//]&lt;component id>[:&lt;version>]</pre>

</center>

Besides dedicated components it is also possible to denote repositories
as a whole:

<center>
    <pre>[+][&lt;type>::][&lt;scheme>:://]&lt;domain>[:&lt;port>][/&lt;repository prefix>]</pre>
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

ghcr.io/mandelsoft/cnudie//github.com/mandelsoft/pause:1.0.0

ctf+tgz::./ctf

```

### SEE ALSO

##### Parents

* [ocm ocm](ocm_ocm.md)	 - Dedicated command flavors for the Open Component Model
* [ocm](ocm.md)	 - Open Component Model command line client

