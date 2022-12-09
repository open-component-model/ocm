// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package comp

import (
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/addhdlrs"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/template"
	common2 "github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/errors"
)

func ProcessComponentDescriptions(ctx clictx.Context, printer common2.Printer, templ template.Options, repo ocm.Repository, h *ResourceSpecHandler, sources []addhdlrs.ElementSource) error {
	elems, ictx, err := addhdlrs.ProcessDescriptions(ctx, printer, templ, h, sources)
	if err != nil {
		return err
	}

	for _, elem := range elems {
		err := h.Add(ctx, ictx.Section("adding %s...", elem.Spec().Info()), elem, repo)
		if err != nil {
			return errors.Wrapf(err, "failed adding component %q(%s)", elem.Spec().GetName(), elem.Source())
		}
	}
	return nil
}
