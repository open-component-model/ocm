// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package hash

import (
	"fmt"
	"strings"

	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/open-component-model/ocm/v2/cmds/ocm/commands/misccmds/names"
	"github.com/open-component-model/ocm/v2/cmds/ocm/commands/verbs"
	"github.com/open-component-model/ocm/v2/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/v2/pkg/contexts/clictx"
	"github.com/open-component-model/ocm/v2/pkg/contexts/ocm/attrs/signingattr"
	"github.com/open-component-model/ocm/v2/pkg/contexts/ocm/compdesc"
	"github.com/open-component-model/ocm/v2/pkg/errors"
	"github.com/open-component-model/ocm/v2/pkg/out"
	"github.com/open-component-model/ocm/v2/pkg/signing"
	"github.com/open-component-model/ocm/v2/pkg/signing/handlers/rsa"
)

var (
	Names = names.Hash
	Verb  = verbs.Create
)

type Command struct {
	utils.BaseCommand

	stype  string
	priv   []byte
	htype  string
	hash   string
	issuer string

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
	}
}

func (o *Command) AddFlags(set *pflag.FlagSet) {
	set.StringVarP(&o.stype, "algorithm", "S", rsa.Algorithm, "signature algorithm")
}

func (o *Command) Complete(args []string) error {
	var err error

	if len(args) < 2 {
		return fmt.Errorf("key file and hash argumnt required")
	}
	if len(args) > 3 {
		return fmt.Errorf("too many arguments")
	}
	if len(args) == 3 {
		o.issuer = args[2]
	}
	o.priv, err = vfs.ReadFile(o.FileSystem(), args[0])
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
	sig, err := o.signer.Sign(o.Context.CredentialsContext(), o.hash, o.hasher.Crypto(), o.issuer, o.priv)
	if err != nil {
		return err
	}
	out.Outf(o, "algorithm: %s\n", sig.Algorithm)
	out.Outf(o, "mediaType: %s\n", sig.MediaType)
	out.Outf(o, "value: %s\n", sig.Value)
	return nil
}
