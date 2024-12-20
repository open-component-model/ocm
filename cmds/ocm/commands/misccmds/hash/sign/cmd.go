package sign

import (
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"strings"

	"github.com/mandelsoft/goutils/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/api/ocm/compdesc"
	"ocm.software/ocm/api/ocm/extensions/attrs/signingattr"
	"ocm.software/ocm/api/tech/signing"
	"ocm.software/ocm/api/tech/signing/handlers/rsa"
	"ocm.software/ocm/api/tech/signing/signutils"
	utils2 "ocm.software/ocm/api/utils"
	"ocm.software/ocm/api/utils/out"
	"ocm.software/ocm/cmds/ocm/commands/misccmds/names"
	"ocm.software/ocm/cmds/ocm/commands/verbs"
	"ocm.software/ocm/cmds/ocm/common/utils"
)

var (
	Names = names.Hash
	Verb  = verbs.Sign
)

type Command struct {
	utils.BaseCommand

	pubFile  string
	rootFile string
	rootCAs  []string

	stype string
	priv  signutils.GenericPrivateKey
	pub   signutils.GenericPublicKey
	roots signutils.GenericCertificatePool
	htype string
	hash  string

	issuer *pkix.Name
	hasher signing.Hasher
	signer signing.Signer
}

var _ utils.OCMCommand = (*Command)(nil)

// NewCommand creates a new artifact command.
func NewCommand(ctx clictx.Context, names ...string) *cobra.Command {
	return utils.SetupCommand(&Command{BaseCommand: utils.NewBaseCommand(ctx)}, utils.Names(Names, names...)...)
}

func (o *Command) ForName(name string) *cobra.Command {
	return &cobra.Command{
		Use:   "<private key file> <hash> [<issuer>]",
		Short: "sign hash",
		Long: `
Print the signature for a dedicated digest value.
	`,
		Example: `
$ ocm sign hash key.priv SHA-256:810ff2fb242a5dee4220f2cb0e6a519891fb67f2f828a6cab4ef8894633b1f50
`,
		Annotations: map[string]string{"ExampleCodeStyle": "bash"},
	}
}

func (o *Command) AddFlags(set *pflag.FlagSet) {
	set.StringVarP(&o.stype, "algorithm", "S", rsa.Algorithm, "signature algorithm")
	set.StringVarP(&o.pubFile, "publicKey", "", "", "public key certificate file")
	set.StringVarP(&o.rootFile, "rootCerts", "", "", "root certificates file (deprecated)")
	set.StringArrayVarP(&o.rootCAs, "ca-cert", "", nil, "additional root certificate authorities (for signing certificates)")
}

func (o *Command) Complete(args []string) error {
	var err error

	if len(args) < 2 {
		return fmt.Errorf("key file and hash argument required")
	}
	if len(args) > 3 {
		return fmt.Errorf("too many arguments")
	}
	if len(args) == 3 {
		o.issuer, err = signutils.ParseDN(args[2])
		if err != nil {
			return errors.Wrapf(err, "issuer")
		}
	}

	if o.pubFile != "" {
		o.pub, err = utils2.ReadFile(o.pubFile, o.FileSystem())
		if err != nil {
			return err
		}
	}

	if o.rootFile != "" {
		roots, err := utils2.ReadFile(o.rootFile, o.FileSystem())
		if err != nil {
			return err
		}
		o.roots, err = signutils.GetCertPool(roots, false)
		if err != nil {
			return err
		}
	}

	if len(o.rootCAs) > 0 {
		var list []*x509.Certificate
		for _, r := range o.rootCAs {
			data, err := utils2.ReadFile(r, o.FileSystem())
			if err != nil {
				return errors.Wrapf(err, "root CA")
			}
			certs, err := signutils.GetCertificateChain(data, false)
			if err != nil {
				return errors.Wrapf(err, "root CA")
			}
			list = append(list, certs...)
		}
		if o.roots != nil {
			for _, c := range list {
				o.roots.(*x509.CertPool).AddCert(c)
			}
		} else {
			o.roots = list
		}
	}

	o.priv, err = utils2.ReadFile(args[0], o.FileSystem())
	if err != nil {
		return err
	}

	if i := strings.Index(args[1], ":"); i <= 0 {
		return fmt.Errorf("hash type missing for hash string")
	} else {
		o.htype = args[1][:i]
		o.hash = args[1][i+1:]
	}

	reg := signingattr.Get(o.Context)
	o.hasher = reg.GetHasher(o.htype)
	if o.hasher == nil {
		return errors.ErrUnknown(compdesc.KIND_HASH_ALGORITHM, o.htype)
	}
	o.signer = reg.GetSigner(o.stype)
	if o.signer == nil {
		return errors.ErrUnknown(compdesc.KIND_SIGN_ALGORITHM, o.stype)
	}
	return nil
}

func (o *Command) Run() error {
	sctx := &signing.DefaultSigningContext{
		Hash:       o.hasher.Crypto(),
		PrivateKey: o.priv,
		PublicKey:  o.pub,
		RootCerts:  o.roots,
		Issuer:     o.issuer,
	}
	sig, err := o.signer.Sign(o.Context.CredentialsContext(), o.hash, sctx)
	if err != nil {
		return err
	}
	out.Outf(o, "algorithm: %s\n", sig.Algorithm)
	out.Outf(o, "mediaType: %s\n", sig.MediaType)
	out.Outf(o, "value: %s\n", sig.Value)
	return nil
}
