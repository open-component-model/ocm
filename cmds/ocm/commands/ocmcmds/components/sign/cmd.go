// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package sign

import (
	"github.com/spf13/cobra"

	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/cmds/signing"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/names"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
)

var (
	Names = names.Components
	Verb  = verbs.Sign
)

// NewCommand creates a new ctf command.
func NewCommand(ctx clictx.Context, names ...string) *cobra.Command {
	return signing.NewCommand(ctx, "Sign", true,
		[]string{"signed", "signing"},
		desc, "$ ocm sign componentversion --signature mandelsoft --private-key=mandelsoft.key ghcr.io/mandelsoft/kubelink",
		utils.Names(Names, names...)...)
}

var desc = ``
