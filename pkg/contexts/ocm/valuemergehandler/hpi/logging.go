package hpi

import (
	ocmlog "github.com/open-component-model/ocm/pkg/logging"
)

var REALM = ocmlog.DefineSubRealm("value marge handling", "valuemerge")

var Log = ocmlog.DynamicLogger(REALM)
