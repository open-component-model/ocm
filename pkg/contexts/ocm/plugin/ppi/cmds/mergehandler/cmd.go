// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package mergehandler

import (
	"github.com/spf13/cobra"

	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/ppi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/ppi/cmds/mergehandler/execute"
)

const Name = "valuemergehandler"

func New(p ppi.Plugin) *cobra.Command {
	cmd := &cobra.Command{
		Use:   Name,
		Short: "value merge handler operations",
		Long:  `This command group provides all commands used to implement an value merge handlers.`,
	}

	cmd.AddCommand(execute.New(p))
	return cmd
}
