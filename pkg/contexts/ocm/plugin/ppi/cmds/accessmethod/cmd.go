// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package accessmethod

import (
	"github.com/spf13/cobra"

	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/ppi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/ppi/cmds/accessmethod/get"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/ppi/cmds/accessmethod/put"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/ppi/cmds/accessmethod/validate"
)

const NAME = "accessmethod"

func New(p ppi.Plugin) *cobra.Command {
	cmd := &cobra.Command{
		Use:   NAME,
		Short: "access method operations",
		Long:  "",
	}

	cmd.AddCommand(validate.New(p))
	cmd.AddCommand(get.New(p))
	cmd.AddCommand(put.New(p))
	return cmd
}
