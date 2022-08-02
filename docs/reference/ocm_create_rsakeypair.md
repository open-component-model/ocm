## ocm create rsakeypair &mdash; Create RSA Public Key Pair

### Synopsis

```
ocm create rsakeypair [<private key file> [<public key file>]] {<subject-attribute>=<value>}
```

### Options

```
      --cacert string       certificate authority to sign public key
      --cakey string        private key for certificate authority
  -h, --help                help for rsakeypair
      --validity duration   certificate validity (default 87600h0m0s)
```

### Description


Create an RSA public key pair and save to files.

The default for the filename to store the private key is <code>rsa.priv</code>.
If no public key file is specified, its name will be derived from the filename for
the private key (suffix <code>.pub</code> for public key or <code>.cert</code> for certificate).
If a certificate authority is given (<code>--cacert</code>) the public key
will be signed. In this case a subject (at least common name/issuer) and a private
key (<code>--cakey</code>) is required. If only a subject is given, the public key will be self-signed.

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

```

$ ocm create rsakeypair mandelsoft.priv mandelsoft.cert issuer=mandelsoft

```

### SEE ALSO

##### Parents

* [ocm create](ocm_create.md)	 &mdash; Create transport or component archive
* [ocm](ocm.md)	 &mdash; Open Component Model command line client

