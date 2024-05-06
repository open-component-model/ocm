package internal

import (
	"github.com/open-component-model/ocm/pkg/contexts/datacontext"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
)

////////////////////////////////////////////////////////////////////////////////

const ATTR_ROUTINGSLIP_ENTRYTYPES = "github.com/open-component-model/ocm/pkg/contexts/ocm/labels/routingslip"

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
