package rsakeypair

import (
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"strings"
	"time"

	"github.com/mandelsoft/goutils/errors"
	parse "github.com/mandelsoft/spiff/dynaml/x509"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/api/datacontext/attrs/rootcertsattr"
	"ocm.software/ocm/api/ocm/extensions/attrs/signingattr"
	"ocm.software/ocm/api/tech/signing"
	"ocm.software/ocm/api/tech/signing/handlers/rsa"
	"ocm.software/ocm/api/tech/signing/signutils"
	utils2 "ocm.software/ocm/api/utils"
	"ocm.software/ocm/api/utils/cobrautils/flag"
	"ocm.software/ocm/api/utils/encrypt"
	"ocm.software/ocm/api/utils/out"
	"ocm.software/ocm/cmds/ocm/commands/misccmds/names"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common"
	"ocm.software/ocm/cmds/ocm/commands/verbs"
	"ocm.software/ocm/cmds/ocm/common/utils"
)

var (
	Names = names.RSAKeyPair
	Verb  = verbs.Create
)

type Command struct {
	utils.BaseCommand

	Subject     *pkix.Name
	MoreIssuers []string
	priv        string
	pub         string
	ekey        string

	attrs     map[string]string
	ca        bool
	rootcerts string
	cacert    string
	cakey     string

	Validity time.Duration

	RootCertPool *x509.CertPool
	CAChain      []*x509.Certificate
	CAKey        interface{}

	Encrypt             string
	CreateEncryptionKey bool
}

var _ utils.OCMCommand = (*Command)(nil)

// NewCommand creates a new artifact command.
func NewCommand(ctx clictx.Context, names ...string) *cobra.Command {
	return utils.SetupCommand(&Command{BaseCommand: utils.NewBaseCommand(ctx)}, utils.Names(Names, names...)...)
}

func (o *Command) ForName(name string) *cobra.Command {
	return &cobra.Command{
		Use:   "[<private key file> [<public key file>]] {<subject-attribute>=<value>}",
		Short: "create RSA public key pair",
		Long: `
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

	`,
		Example: `
$ ocm create rsakeypair mandelsoft.priv mandelsoft.cert issuer=mandelsoft
`,
		Annotations: map[string]string{"ExampleCodeStyle": "bash"},
	}
}

func (o *Command) AddFlags(set *pflag.FlagSet) {
	set.BoolVarP(&o.ca, "ca", "", false, "create certificate for a signing authority")
	set.StringVarP(&o.rootcerts, "root-certs", "", "", "root certificates used to validate used certificate authority")
	set.StringVarP(&o.cacert, "ca-cert", "", "", "certificate authority to sign public key")
	set.StringVarP(&o.cakey, "ca-key", "", "", "private key for certificate authority")
	set.DurationVarP(&o.Validity, "validity", "", 10*24*365*time.Hour, "certificate validity")
	set.StringVarP(&o.Encrypt, "encryptionKey", "e", "", "encrypt private key with given key")
	set.BoolVarP(&o.CreateEncryptionKey, "encrypt", "E", false, "encrypt private key with new key")

	flag.StringVarPF(set, &o.cacert, "cacert", "", "", "certificate authority to sign public key").Hidden = true
	flag.StringVarPF(set, &o.cakey, "cakey", "", "", "private key for certificate authority").Hidden = true
}

func (o *Command) FilterSettings(args ...string) []string {
	o.attrs, args = common.FilterSettings(args...)
	return args
}

func (o *Command) Complete(args []string) error {
	args = o.FilterSettings(args...)

	if len(args) > 2 {
		return errors.Newf("only a maximum of two filenames possible")
	}
	if o.CreateEncryptionKey && o.Encrypt != "" {
		return errors.Newf("only one of --encrypt or --encryptionKey is possible")
	}

	if o.rootcerts != "" {
		pool, err := signutils.GetCertPool(o.rootcerts, false)
		if err != nil {
			path, err := utils2.ResolvePath(o.rootcerts)
			if err != nil {
				return errors.Wrapf(err, "failed to resolve root certificates")
			}
			data, err := vfs.ReadFile(o.Context.FileSystem(), path)
			if err != nil {
				return errors.Wrapf(err, "cannot read root cert file %q", o.rootcerts)
			}
			pool, err = signutils.GetCertPool(data, false)
			if err != nil {
				return errors.Wrapf(err, "no root cert in file %q", o.rootcerts)
			}
		}
		o.RootCertPool = pool
	} else {
		o.RootCertPool = rootcertsattr.Get(o.Context).GetRootCertPool(true)
	}

	if len(o.attrs) > 0 {
		var subject pkix.Name
		for k, v := range o.attrs {
			switch strings.ToLower(k) {
			case "issuer", "common-name", "cn":
				if subject.CommonName == "" {
					subject.CommonName = v
				} else {
					return fmt.Errorf("issuer already set")
				}
			case "street":
				subject.StreetAddress = append(subject.StreetAddress, v)
			case "province", "st":
				subject.Province = append(subject.Province, v)
			case "country", "c":
				subject.Country = append(subject.Country, v)
			case "organization", "org", "o":
				subject.Country = append(subject.Country, v)
			case "organizational-unit", "org-unit", "ou":
				subject.OrganizationalUnit = append(subject.OrganizationalUnit, v)
			case "postal-code", "postalcode":
				subject.PostalCode = append(subject.PostalCode, v)
			case "locality", "l":
				subject.Locality = append(subject.Locality, v)
			default:
				return errors.ErrUnknown("subject attribute,", k)
			}
		}
		o.Subject = &subject
	}

	if o.cacert != "" {
		raw := []byte(o.cacert)
		cert, pool, err := signutils.GetCertificate(o.cacert, false)
		if err != nil {
			path, err := utils2.ResolvePath(o.cacert)
			if err != nil {
				return errors.Wrapf(err, "failed to resolve cacert file %q", o.cacert)
			}
			data, err := vfs.ReadFile(o.Context.FileSystem(), path)
			if err != nil {
				return errors.Wrapf(err, "cannot read ca cert file %q", o.cacert)
			}
			cert, pool, err = signutils.GetCertificate(data, false)
			if err != nil {
				return errors.Wrapf(err, "no cert in file %q", o.cacert)
			}
			raw = data
		}

		if !signutils.IsSelfSigned(cert) {
			opts := x509.VerifyOptions{
				Intermediates: pool,
				Roots:         o.RootCertPool,
				CurrentTime:   time.Time{},
				KeyUsages:     []x509.ExtKeyUsage{x509.ExtKeyUsageCodeSigning},
			}
			_, err = cert.Verify(opts)
			if err != nil {
				return err
			}
		}
		o.CAChain, err = signutils.GetCertificateChain(raw, false)
		if err != nil {
			return err
		}
	}
	if o.cakey != "" {
		key, err := parse.ParsePrivateKey(o.cakey)
		if err != nil {
			path, err := utils2.ResolvePath(o.cakey)
			if err != nil {
				return errors.Wrapf(err, "failed to resolve ca key file %q", o.cakey)
			}
			data, err := vfs.ReadFile(o.Context.FileSystem(), path)
			if err != nil {
				return errors.Wrapf(err, "cannot read private key file %q", o.cakey)
			}
			key, err = parse.ParsePrivateKey(string(data))
			if err != nil {
				return errors.Wrapf(err, "unknown private key in file %q", o.cakey)
			}
		}
		o.CAKey = key
	}
	if len(o.CAChain) != 0 && o.CAKey == nil {
		return errors.Newf("private key required for signing public key")
	}
	if len(o.CAChain) == 0 && o.CAKey != nil {
		return errors.Newf("ca certificate required for signing public key")
	}

	if o.Subject != nil {
		if o.Subject.CommonName == "" {
			return errors.Newf("at least the common-name for a subject must be given")
		}
	}
	if len(args) > 0 {
		o.priv = args[0]
	} else {
		o.priv = "rsa.priv"
	}
	if len(args) > 1 {
		o.pub = args[1]
	} else {
		suf := "pub"
		if o.Subject != nil {
			suf = "cert"
		}
		if strings.HasSuffix(o.priv, ".priv") {
			o.pub = o.priv[:len(o.priv)-4] + suf
		} else {
			o.pub = o.priv + "." + suf
		}
	}

	if o.ekey == "" {
		o.ekey = o.priv + ".ekey"
	}

	return nil
}

func (o *Command) Run() error {
	raw := false

	priv, pub, err := rsa.Handler{}.CreateKeyPair()
	if err != nil {
		return err
	}

	if o.Subject != nil {
		key := o.CAKey
		if o.CAKey == nil {
			key = priv
		}

		spec := &signutils.Specification{
			RootCAs:      o.RootCertPool,
			IsCA:         o.ca,
			PublicKey:    pub,
			CAPrivateKey: key,
			CAChain:      o.CAChain,
			Subject:      *o.Subject,
			Usages:       signutils.Usages{x509.ExtKeyUsageCodeSigning},
			Validity:     o.Validity,
			NotBefore:    nil,
		}
		if len(o.CAChain) == 1 && signutils.IsSelfSigned(o.CAChain[0]) {
			o.RootCertPool.AddCert(o.CAChain[0])
		}

		_, pub, err = signutils.CreateCertificate(spec)
		if err != nil {
			return errors.Wrapf(err, "signing of key pair failed")
		}
		raw = true
	}

	var key []byte
	if o.CreateEncryptionKey {
		key, err = encrypt.NewKey(encrypt.AES_256)
		if err != nil {
			return errors.Wrapf(err, "cannot create new encryption key")
		}
	}
	if o.Encrypt != "" {
		reg := signingattr.Get(o.Context.OCMContext())
		p, err := signing.ResolvePrivateKey(reg, signing.DecryptionKeyName(o.Encrypt))
		if err != nil {
			return err
		}
		key, err = encrypt.KeyFromAny(p)
		if err != nil {
			return errors.Wrapf(err, "key %q", signing.DecryptionKeyName(o.Encrypt))
		}
	}
	if key != nil {
		data, err := rsa.KeyData(priv)
		if err != nil {
			return err
		}
		algo, err := encrypt.AlgoForKey(key)
		if err != nil {
			return errors.Wrapf(err, "key %q", signing.DecryptionKeyName(o.Encrypt))
		}
		cipherText, err := encrypt.Encrypt(key, data)
		if err != nil {
			return err
		}
		priv = encrypt.EncryptedToPem(algo, cipherText)
		if o.CreateEncryptionKey {
			if err := o.WriteKey(encrypt.KeyToPem(key), o.ekey, true); err != nil {
				return errors.Wrapf(err, "failed to write encryption key file %q", o.ekey)
			}
		}
	}
	if err := o.WriteKey(priv, o.priv, key != nil); err != nil {
		return errors.Wrapf(err, "failed to write private key file %q", o.priv)
	}
	if err := o.WriteKey(pub, o.pub, raw); err != nil {
		return errors.Wrapf(err, "failed to write public key file %q", o.pub)
	}
	msg := ""
	add := ""
	if key != nil {
		msg = " encrypted"
		if o.CreateEncryptionKey {
			add = "[" + o.ekey + "]"
		}
	}
	out.Outf(o.Context, "created%s rsa key pair %s[%s]%s\n", msg, o.priv, o.pub, add)
	return nil
}

func (o *Command) WriteKey(key interface{}, path string, raw bool) error {
	fd, err := o.Context.FileSystem().OpenFile(path, vfs.O_CREATE|vfs.O_WRONLY, 0o600)
	if err != nil {
		return err
	}

	if certdata, ok := key.([]byte); ok {
		if raw {
			_, err = fd.Write(certdata)
		} else {
			block := &pem.Block{Type: "CERTIFICATE", Bytes: certdata}
			err = pem.Encode(fd, block)
		}
	} else {
		err = rsa.WriteKeyData(key, fd)
	}
	if err != nil {
		fd.Close()
		o.Context.FileSystem().Remove(path)
		return err
	}
	err = fd.Close()
	if err != nil {
		return err
	}
	return o.FileSystem().Chmod(path, 0o400)
}
