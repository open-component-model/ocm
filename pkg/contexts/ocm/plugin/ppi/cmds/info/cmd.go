// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package info

import (
	"encoding/json"

	"github.com/spf13/cobra"

	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/ppi"
)

const NAME = "info"

func New(p ppi.Plugin) *cobra.Command {
	return &cobra.Command{
		Use:   NAME,
		Short: "show plugin descriptor",
		Long:  "",
		Args:  cobra.MaximumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			data, err := json.Marshal(p.Descriptor())
			if err != nil {
				return err
			}
			cmd.Printf("%s\n", string(data))
			return nil
		},
	}
}
