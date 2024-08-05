package internal

import (
	"ocm.software/ocm/api/datacontext"
	"ocm.software/ocm/api/ocm/cpi"
)

////////////////////////////////////////////////////////////////////////////////

const ATTR_ROUTINGSLIP_ENTRYTYPES = "ocm.software/ocm/api/ocm/extensions/labels/routingslip"

func For(ctx cpi.ContextProvider) EntryTypeScheme {
	if ctx == nil {
		return DefaultEntryTypeScheme()
	}
	return ctx.OCMContext().GetAttributes().GetOrCreateAttribute(ATTR_ROUTINGSLIP_ENTRYTYPES, create).(EntryTypeScheme)
}

func create(datacontext.Context) interface{} {
	return NewEntryTypeScheme(DefaultEntryTypeScheme())
}

func SetFor(ctx datacontext.Context, registry EntryTypeScheme) {
	ctx.GetAttributes().SetAttribute(ATTR_ROUTINGSLIP_ENTRYTYPES, registry)
}
