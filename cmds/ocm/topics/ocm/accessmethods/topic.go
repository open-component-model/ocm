// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package topicocmaccessmethods

import (
	"github.com/spf13/cobra"

	"github.com/open-component-model/ocm/pkg/contexts/clictx"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
)

func New(ctx clictx.Context) *cobra.Command {
	return &cobra.Command{
		Use:   "ocm-accessmethods",
		Short: "List of all supported access methods",

		Long: `
Access methods are used to handle the access to the content of artefacts
described in a component version. Therefore, an artefact entry contains
an access specification describing the access attributes for the dedicated
artefact.

` + ocm.AccessUsage(ctx.OCMContext().AccessMethods(), true),
	}
}
