## ocm sign componentversions &mdash; Sign Component Version

### Synopsis

```bash
ocm sign componentversions [<options>] {<component-reference>}
```

#### Aliases

```text
componentversions, componentversion, cv, components, component, comps, comp, c
```

### Options

```text
      --                          enable verification store
  -S, --algorithm string          signature handler (default "RSASSA-PKCS1-V1_5")
      --ca-cert stringArray       additional root certificate authorities (for signing certificates)
  -c, --constraints constraints   version constraint
  -H, --hash string               hash algorithm (default "SHA-256")
  -h, --help                      help for componentversions
  -I, --issuer stringArray        issuer name or distinguished name (DN) (optionally for dedicated signature) ([<name>:=]<dn>)
      --keyless                   use keyless signing
      --latest                    restrict component versions to latest
      --lookup stringArray        repository name or spec for closure lookup fallback
  -N, --normalization string      normalization algorithm (default "jsonNormalisation/v3")
  -K, --private-key stringArray   private key setting
  -k, --public-key stringArray    public key setting
  -R, --recursive                 recursively sign component versions
      --repo string               repository name or spec
  -s, --signature stringArray     signature name
      --tsa                       use timestamp authority (default server: http://timestamp.digicert.com)
      --tsa-url string            TSA server URL
      --update                    update digest in component versions (default true)
      --verified string           file used to remember verifications for downloads (default "~/.ocm/verified")
  -V, --verify                    verify existing digests (default true)
```

### Description

Sign specified component versions.


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


The <code>--public-key</code> and <code>--private-key</code> options can be
used to define public and private keys on the command line. The options have an
argument of the form <code>[&lt;name>=]&lt;filepath></code>. The optional name
specifies the signature name the key should be used for. By default, this is the
signature name specified with the option <code>--signature</code>.

Alternatively a key can be specified as base64 encoded string if the argument
start with the prefix <code>!</code> or as direct string with the prefix
<code>=</code>.

If the verification store is enabled, resources downloaded from
signed or verified component versions are verified against their digests
provided by the component version.(not supported for using downloaders for the
resource download).

The usage of the verification store is enabled by <code>--</code> or by
specifying a verification file with <code>--verified</code>.

If in signing mode a public key is specified, existing signatures for the
given signature name will be verified, instead of recreated.


The following signing types are supported with option <code>--algorithm</code>:
  - <code>RSASSA-PKCS1-V1_5</code> (default)
  - <code>RSASSA-PSS</code>
  - <code>rsa-signingservice</code>
  - <code>rsapss-signingservice</code>
  - <code>sigstore</code>
  - <code>sigstore-v3</code>


The following normalization modes are supported with option <code>--normalization</code>:
  - <code>jsonNormalisation/v1</code>
  - <code>jsonNormalisation/v2</code>
  - <code>jsonNormalisation/v3</code> (default)


The following hash modes are supported with option <code>--hash</code>:
  - <code>NO-DIGEST</code>
  - <code>SHA-256</code> (default)
  - <code>SHA-512</code>

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
$ ocm sign componentversion --signature mysignature --private-key=my.key ghcr.io/open-component-model/ocm//ocm.software/ocmcli:0.17.0
```

### SEE ALSO

#### Parents

* [ocm sign](ocm_sign.md)	 &mdash; Sign components or hashes
* [ocm](ocm.md)	 &mdash; Open Component Model command line client

