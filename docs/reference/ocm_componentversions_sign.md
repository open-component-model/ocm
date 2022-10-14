## ocm componentversions sign &mdash; Sign Component Version

### Synopsis

```
ocm componentversions sign [<options>] {<component-reference>}
```

### Options

```
  -S, --algorithm string          signature handler (default "RSASSA-PKCS1-V1_5")
      --ca-cert stringArray       Additional root certificates
  -H, --hash string               hash algorithm (default "sha256")
  -h, --help                      help for sign
  -I, --issuer string             issuer name
      --lookup stringArray        repository name or spec for closure lookup fallback
  -N, --normalization string      normalization algorithm (default "jsonNormalisation/v1")
  -K, --private-key stringArray   private key setting
  -k, --public-key stringArray    public key setting
  -R, --recursive                 recursively sign component versions
      --repo string               repository name or spec
  -s, --signature stringArray     signature name
      --update                    update digest in component versions (default true)
  -V, --verify                    verify existing digests (default true)
```

### Description


Sign specified component versions. 

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

Using the JSON variant any repository type supported by the 
linked library can be used:

Dedicated OCM repository types:
- `ComponentArchive`

OCI Repository types (using standard component repository to OCI mapping):
- `ArtefactSet`
- `CommonTransportFormat`
- `DockerDaemon`
- `Empty`
- `OCIRegistry`
- `oci`
- `ociRegistry`

The <code>--public-key</code> and <code>--private-key</code> options can be
used to define public and private keys on the command line. The options have an
argument of the form <code>[&lt;name>=]&lt;filepath></code>. The optional name
specifies the signature name the key should be used for. By default this is the
signature name specified with the option <code>--signature</code>.

Alternatively a key can be specified as base64 encoded string if the argument
start with the prefix <code>!</code> or as direct string with the prefix
<code>=</code>.

If in signing mode a public key is specified, existing signatures for the
given signature name will be verified, instead of recreated.


The following signing types are supported with option <code>--algorithm</code>:

  - <code>RSASSA-PKCS1-V1_5</code> (default): 

  - <code>rsa-signingsservice</code>: 



The following normalization modes are supported with option <code>--normalization</code>:

  - <code>jsonNormalisation/v1</code> (default): 

  - <code>jsonNormalisation/v2</code>: 



The following hash modes are supported with option <code>--hash</code>:

  - <code>NO-DIGEST</code>: 

  - <code>sha256</code> (default): 

  - <code>sha512</code>: 


If a component lookup for building a reference closure is required
the <code>--lookup</code>  option can be used to specify a fallback
lookup repository. 
By default the component versions are searched in the repository
holding the component version for which the closure is determined.
For *Component Archives* this is never possible, because it only
contains a single component version. Therefore, in this scenario
this option must always be specified to be able to follow component
references.


### Examples

```
$ ocm sign componentversion --signature mandelsoft --private-key=mandelsoft.key ghcr.io/mandelsoft/kubelink
```

### SEE ALSO

##### Parents

* [ocm componentversions](ocm_componentversions.md)	 &mdash; Commands acting on components
* [ocm](ocm.md)	 &mdash; Open Component Model command line client

