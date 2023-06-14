// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package accessmethod

import (
	"github.com/spf13/cobra"

	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/ppi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/ppi/cmds/accessmethod/compose"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/ppi/cmds/accessmethod/get"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/ppi/cmds/accessmethod/identity"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/ppi/cmds/accessmethod/validate"
)

const Name = "accessmethod"

func New(p ppi.Plugin) *cobra.Command {
	cmd := &cobra.Command{
		Use:   Name,
		Short: "access method operations",
		Long: `This command group provides all commands used to implement an access method
described by an access method descriptor (<CMD>` + p.Name() + ` descriptor</CMD>.`,
	}

	cmd.AddCommand(validate.New(p))
	cmd.AddCommand(get.New(p))
	cmd.AddCommand(identity.New(p))
	cmd.AddCommand(compose.New(p))
	return cmd
}
