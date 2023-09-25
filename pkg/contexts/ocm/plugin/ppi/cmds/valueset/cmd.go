// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package valueset

import (
	"github.com/spf13/cobra"

	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/ppi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/ppi/cmds/valueset/compose"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/ppi/cmds/valueset/validate"
)

const Name = "valueset"

func New(p ppi.Plugin) *cobra.Command {
	cmd := &cobra.Command{
		Use:   Name,
		Short: "valueset operations",
		Long: `This command group provides all commands used to implement a a value set
described by a value set descriptor (<CMD>` + p.Name() + ` descriptor</CMD>.`,
	}

	cmd.AddCommand(compose.New(p))
	cmd.AddCommand(validate.New(p))
	return cmd
}
