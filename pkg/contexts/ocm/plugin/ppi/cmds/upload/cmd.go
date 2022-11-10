// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package upload

import (
	"github.com/spf13/cobra"

	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/ppi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/ppi/cmds/upload/put"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/ppi/cmds/upload/validate"
)

const Name = "upload"

func New(p ppi.Plugin) *cobra.Command {
	cmd := &cobra.Command{
		Use:   Name,
		Short: "upload specific operations",
		Long: `
This command group provides all commands used to implement an uploader
described by an uploader descriptor.`,
	}

	cmd.AddCommand(validate.New(p))
	cmd.AddCommand(put.New(p))
	return cmd
}
