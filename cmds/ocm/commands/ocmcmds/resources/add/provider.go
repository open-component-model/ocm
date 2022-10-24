//  SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
//  SPDX-License-Identifier: Apache-2.0

package add

import (
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common"
	"github.com/open-component-model/ocm/pkg/cobrautils/flagsets"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
)

type ResourceSpecificationsProvider struct {
	*common.ContentResourceSpecificationsProvider
}

func NewResourceSpecificationsProvider(ctx clictx.Context, deftype string) common.ResourceSpecificationsProvider {
	a := &ResourceSpecificationsProvider{}
	a.ContentResourceSpecificationsProvider = common.NewContentResourceSpecificationProvider(ctx, "resource", a.addMeta, deftype,
		flagsets.NewBoolOptionType("external", "flag non-local resource"),
	)
	return a
}

func (p *ResourceSpecificationsProvider) addMeta(opts flagsets.ConfigOptions, config flagsets.Config) error {
	if o, ok := opts.GetValue("external"); ok && o.(bool) {
		config["relation"] = metav1.ExternalRelation
	}
	return nil
}

func (p *ResourceSpecificationsProvider) Description() string {
	d := p.ContentResourceSpecificationsProvider.Description()
	return d + "Non-local resources can be indicated using the option <code>--external</code>."
}
