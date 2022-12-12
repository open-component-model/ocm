// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package comp

import (
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/addhdlrs"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/inputs"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/errors"
)

func ProcessComponents(ctx clictx.Context, ictx inputs.Context, repo ocm.Repository, h *ResourceSpecHandler, elems []addhdlrs.Element) error {
	for _, elem := range elems {
		err := h.Add(ctx, ictx.Section("adding %s...", elem.Spec().Info()), elem, repo)
		if err != nil {
			return errors.Wrapf(err, "failed adding component %q(%s)", elem.Spec().GetName(), elem.Source())
		}
	}
	return nil
}
