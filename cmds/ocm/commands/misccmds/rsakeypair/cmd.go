// Copyright 2022 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package rsakeypair

import (
	"strings"

	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/open-component-model/ocm/cmds/ocm/commands/misccmds/names"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/out"
	"github.com/open-component-model/ocm/pkg/signing/handlers/rsa"
	"github.com/spf13/cobra"

	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"

	"github.com/open-component-model/ocm/cmds/ocm/clictx"
)

var (
	Names = names.RSAKeyPair
	Verb  = verbs.Create
)

type Command struct {
	utils.BaseCommand

	priv string
	pub  string
}

// NewCommand creates a new artefact command.
func NewCommand(ctx clictx.Context, names ...string) *cobra.Command {
	return utils.SetupCommand(&Command{BaseCommand: utils.NewBaseCommand(ctx)}, utils.Names(Names, names...)...)
}

func (o *Command) ForName(name string) *cobra.Command {
	return &cobra.Command{
		Use:   "[<private key file> [<public key file]]",
		Short: "create RSA public key pair",
		Long: `
Create an RSA public key pair and save to files.
	`,
		Args: cobra.MaximumNArgs(2),
		Example: `
$ ocm create rsakeypair mandelsoft.priv mandelsoft.pub
`,
	}
}

func (o *Command) Complete(args []string) error {
	if len(args) > 0 {
		o.priv = args[0]
	} else {
		o.priv = "rsa.priv"
	}
	if len(args) > 1 {
		o.pub = args[1]
	} else {
		if strings.HasSuffix(o.priv, ".priv") {
			o.pub = o.priv[:len(o.priv)-4] + "pub"
		} else {
			o.pub = o.priv + ".pub"
		}
	}
	return nil
}

func (o *Command) Run() error {

	priv, pub, err := rsa.Handler{}.CreateKeyPair()
	if err != nil {
		return err
	}

	if err := o.WriteKey(priv, o.priv); err != nil {
		return errors.Wrapf(err, "failed to write private key file %q", o.priv)
	}
	if err := o.WriteKey(pub, o.pub); err != nil {
		return errors.Wrapf(err, "failed to write public key file %q", o.pub)
	}
	out.Outf(o.Context, "created rsa key pair %s[%s]\n", o.priv, o.pub)
	return nil
}

func (o *Command) WriteKey(key interface{}, path string) error {
	fd, err := o.Context.FileSystem().OpenFile(path, vfs.O_CREATE|vfs.O_WRONLY, 0600)
	if err != nil {
		return err
	}

	err = rsa.WriteKeyData(key, fd)
	if err != nil {
		fd.Close()
		o.Context.FileSystem().Remove(path)
		return err
	}
	err = fd.Close()
	if err != nil {
		return err
	}
	return o.FileSystem().Chmod(path, 0400)
}
