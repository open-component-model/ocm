---
title: "ocm sign hash &mdash; Sign Hash"
url: "/docs/cli-reference/sign/hash/"
sidebar:
  collapsed: true
---

### Synopsis

```bash
ocm sign hash <private key file> <hash> [<issuer>]
```

### Options

```text
  -S, --algorithm string      signature algorithm (default "RSASSA-PKCS1-V1_5")
      --ca-cert stringArray   additional root certificate authorities (for signing certificates)
  -h, --help                  help for hash
      --publicKey string      public key certificate file
      --rootCerts string      root certificates file (deprecated)
```

### Description

Print the signature for a dedicated digest value.
	
### Examples

```bash
$ ocm sign hash key.priv SHA-256:810ff2fb242a5dee4220f2cb0e6a519891fb67f2f828a6cab4ef8894633b1f50
```

### SEE ALSO

#### Parents

* [ocm sign](ocm_sign.md)	 &mdash; Sign components or hashes
* [ocm](ocm.md)	 &mdash; Open Component Model command line client

