package keyoption

import (
	"crypto/x509"
	"fmt"
	"reflect"
	"strings"

	"github.com/mandelsoft/goutils/errors"
	"github.com/spf13/pflag"

	"ocm.software/ocm/api/datacontext"
	"ocm.software/ocm/api/datacontext/attrs/vfsattr"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/tech/signing"
	"ocm.software/ocm/api/tech/signing/signutils"
	"ocm.software/ocm/api/utils"
)

type ConfigFragment struct {
	DefaultName string   `json:"defaultName,omitempty"`
	PublicKeys  []string `json:"publicKeys,omitempty"`
	PrivateKeys []string `json:"privateKeys,omitempty"`
	Issuers     []string `json:"issuers,omitempty"`
	RootCAs     []string `json:"rootCAs,omitempty"`
}

func (c *ConfigFragment) AddFlags(fs *pflag.FlagSet) {
	fs.StringArrayVarP(&c.PublicKeys, "public-key", "k", nil, "public key setting")
	fs.StringArrayVarP(&c.PrivateKeys, "private-key", "K", nil, "private key setting")
	fs.StringArrayVarP(&c.Issuers, "issuer", "I", nil, "issuer name or distinguished name (DN) (optionally for dedicated signature) ([<name>:=]<dn>)")
	fs.StringArrayVarP(&c.RootCAs, "ca-cert", "", nil, "additional root certificate authorities (for signing certificates)")
}

func (c *ConfigFragment) Evaluate(ctx ocm.Context, keys signing.KeyRegistry) (*EvaluatedOptions, error) {
	var opts EvaluatedOptions

	if keys == nil {
		keys = signing.NewKeyRegistry()
	}
	opts.Keys = keys

	err := c.HandleKeys(ctx, "public key", c.PublicKeys, keys.RegisterPublicKey)
	if err != nil {
		return nil, err
	}
	err = c.HandleKeys(ctx, "private key", c.PrivateKeys, keys.RegisterPrivateKey)
	if err != nil {
		return nil, err
	}
	for _, i := range c.Issuers {
		name := c.DefaultName
		is := i
		sep := strings.Index(i, ":=")
		if sep >= 0 {
			name = i[:sep]
			is = i[sep+1:]
		}
		old := keys.GetIssuer(name)
		dn, err := signutils.ParseDN(is)
		if err != nil {
			return nil, errors.Wrapf(err, "issuer %q", i)
		}
		if old != nil && !reflect.DeepEqual(old, dn) {
			return nil, fmt.Errorf("issuer already set (%s)", i)
		}

		keys.RegisterIssuer(name, dn)
	}

	if len(c.RootCAs) > 0 {
		var list []*x509.Certificate
		for _, r := range c.RootCAs {
			data, err := utils.ReadFile(r, vfsattr.Get(ctx))
			if err != nil {
				return nil, errors.Wrapf(err, "root CA")
			}
			certs, err := signutils.GetCertificateChain(data, false)
			if err != nil {
				return nil, errors.Wrapf(err, "root CA")
			}
			list = append(list, certs...)
		}
		opts.RootCerts = list
	}
	return &opts, nil
}

func (c *ConfigFragment) HandleKeys(ctx datacontext.Context, desc string, keys []string, add func(string, interface{})) error {
	name := c.DefaultName
	fs := vfsattr.Get(ctx)
	for _, k := range keys {
		file := k
		sep := strings.Index(k, "=")
		if sep > 0 {
			name = k[:sep]
			file = k[sep+1:]
		}
		if len(file) == 0 {
			return errors.Newf("%s: empty file name", desc)
		}
		var data []byte
		var err error
		switch file[0] {
		case '=', '!', '@':
			data, err = utils.ResolveData(file, fs)
		default:
			data, err = utils.ReadFile(file, fs)
		}
		if err != nil {
			return errors.Wrapf(err, "cannot read %s file %q", desc, file)
		}
		if name == "" {
			return errors.Newf("%s: key name required", desc)
		}
		add(name, data)
	}
	return nil
}
