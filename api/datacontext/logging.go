package datacontext

import (
	ocmlog "ocm.software/ocm/api/utils/logging"
)

var Realm = ocmlog.DefineSubRealm("context lifecycle", "context")

var Logger = ocmlog.DynamicLogger(Realm)

func Debug(c Context, msg string, keypairs ...interface{}) {
	c.LoggingContext().Logger(Realm).Debug(msg, append(keypairs, "id", c.GetId())...)
}
