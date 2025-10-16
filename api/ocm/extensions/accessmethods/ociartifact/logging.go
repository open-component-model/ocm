package ociartifact

import (
	"github.com/mandelsoft/logging"
	"ocm.software/ocm/api/ocm/cpi"
	ocmlog "ocm.software/ocm/api/utils/logging"
)

var REALM = ocmlog.DefineSubRealm("access method ociArtifact", "accessmethod/ociartifact")

type ContextProvider interface {
	GetContext() cpi.Context
}

func Logger(c ContextProvider, keyValuePairs ...interface{}) logging.Logger {
	return c.GetContext().Logger(REALM).WithValues(keyValuePairs...)
}

type localContextProvider struct {
	cpi.ContextProvider
}

func (l *localContextProvider) GetContext() cpi.Context {
	return l.OCMContext()
}

func WrapContextProvider(ctx cpi.ContextProvider) ContextProvider {
	return &localContextProvider{ctx}
}
