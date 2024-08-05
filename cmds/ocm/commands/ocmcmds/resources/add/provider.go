package add

import (
	clictx "ocm.software/ocm/api/cli"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/utils/cobrautils/flagsets"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common"
)

type ResourceSpecificationsProvider struct {
	*common.ContentResourceSpecificationsProvider
}

func NewResourceSpecificationsProvider(ctx clictx.Context, deftype string) common.ElementSpecificationsProvider {
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
	return d + "Non-local resources can be indicated using the option <code>--external</code>.\n"
}
