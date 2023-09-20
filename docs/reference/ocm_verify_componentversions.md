## ocm verify componentversions &mdash; Verify Signature Of Component Version

### Synopsis

```
ocm verify componentversions [<options>] {<component-reference>}
```

##### Aliases

```
componentversions, componentversion, cv, components, component, comps, comp, c
```

### Options

```
      --ca-cert stringArray       additional root certificates
  -c, --constraints constraints   version constraint
  -h, --help                      help for componentversions
      --keyless                   use keyless signing
      --latest                    restrict component versions to latest
  -L, --local                     verification based on information found in component versions, only
      --lookup stringArray        repository name or spec for closure lookup fallback
  -K, --private-key stringArray   private key setting
  -k, --public-key stringArray    public key setting
      --repo string               repository name or spec
  -s, --signature stringArray     signature name
  -V, --verify                    verify existing digests
```

### Description


Verify signature of specified component versions.


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

Dedicated OCM repository types:
  - <code>ComponentArchive</code>: v1

OCI Repository types (using standard component repository to OCI mapping):
  - <code>ArtifactSet</code>: v1
  - <code>CommonTransportFormat</code>: v1
  - <code>DockerDaemon</code>: v1
  - <code>Empty</code>: v1
  - <code>OCIRegistry</code>: v1
  - <code>oci</code>: v1
  - <code>ociRegistry</code>


The <code>--public-key</code> and <code>--private-key</code> options can be
used to define public and private keys on the command line. The options have an
argument of the form <code>[&lt;name>=]&lt;filepath></code>. The optional name
specifies the signature name the key should be used for. By default, this is the
signature name specified with the option <code>--signature</code>.

Alternatively a key can be specified as base64 encoded string if the argument
start with the prefix <code>!</code> or as direct string with the prefix
<code>=</code>.

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

```
$ ocm verify componentversion --signature mandelsoft --public-key=mandelsoft.key ghcr.io/mandelsoft/kubelink
```

### SEE ALSO

##### Parents

* [ocm verify](ocm_verify.md)	 &mdash; Verify component version signatures
* [ocm](ocm.md)	 &mdash; Open Component Model command line client

