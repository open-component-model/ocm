---
title: "create rsakeypair"
url: "/docs/cli-reference/create/rsakeypair/"
---

## ocm create rsakeypair &mdash; Create RSA Public Key Pair

### Synopsis

```bash
ocm create rsakeypair [<private key file> [<public key file>]] {<subject-attribute>=<value>}
```

#### Aliases

```text
rsakeypair, rsa
```

### Options

```text
      --ca                     create certificate for a signing authority
      --ca-cert string         certificate authority to sign public key
      --ca-key string          private key for certificate authority
  -E, --encrypt                encrypt private key with new key
  -e, --encryptionKey string   encrypt private key with given key
  -h, --help                   help for rsakeypair
      --root-certs string      root certificates used to validate used certificate authority
      --validity duration      certificate validity (default 87600h0m0s)
```

### Description

Create an RSA public key pair and save to files.

The default for the filename to store the private key is <code>rsa.priv</code>.
If no public key file is specified, its name will be derived from the filename for
the private key (suffix <code>.pub</code> for public key or <code>.cert</code>
for certificate). If a certificate authority is given (<code>--ca-cert</code>)
the public key will be signed. In this case a subject (at least common
name/issuer) and a private key (<code>--ca-key</code>) for the ca used to sign the
key is required.

If only a subject is given and no ca, the public key will be self-signed.
A signed public key always contains the complete certificate chain. If a
non-self-signed ca is used to sign the key, its certificate chain is verified.
Therefore, an additional root certificate (<code>--root-certs</code>) is required,
if no public root certificate was used to create the used ca.

For signing the public key the following subject attributes are supported:
- <code>CN</code>, <code>common-name</code>, <code>issuer</code>: Common Name/Issuer
- <code>O</code>, <code>organization</code>, <code>org</code>: Organization
- <code>OU</code>, <code>organizational-unit</code>, <code>org-unit</code>: Organizational Unit
- <code>STREET</code> (multiple): Street Address
- <code>POSTALCODE</code>, <code>postal-code</code> (multiple): Postal Code
- <code>L</code>, <code>locality</code> (multiple): Locality
- <code>S</code>, <code>province</code>, (multiple): Province
- <code>C</code>, <code>country</code>, (multiple): Country

	
### Examples

```bash
$ ocm create rsakeypair mandelsoft.priv mandelsoft.cert issuer=mandelsoft
```

### SEE ALSO

#### Parents

* [ocm create](ocm_create.md)	 &mdash; Create transport or component archive
* [ocm](ocm.md)	 &mdash; Open Component Model command line client

