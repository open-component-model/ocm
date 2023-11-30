## ocm sign hash &mdash; Sign Hash

### Synopsis

```
ocm sign hash <private key file> <hash> [<issuer>]
```

### Options

```
  -S, --algorithm string   signature algorithm (default "RSASSA-PKCS1-V1_5")
  -h, --help               help for hash
      --publicKey string   public key certificate file
      --rootCerts string   root certificates file
```

### Description


Print the signature for a dedicated digest value.
	

### Examples

```
$ ocm sign hash key.priv SHA-256:810ff2fb242a5dee4220f2cb0e6a519891fb67f2f828a6cab4ef8894633b1f50
```

### SEE ALSO

##### Parents

* [ocm sign](ocm_sign.md)	 &mdash; Sign components or hashes
* [ocm](ocm.md)	 &mdash; Open Component Model command line client

