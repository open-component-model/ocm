package stdopts

import (
	"ocm.software/ocm/api/datacontext/attrs/tmpcache"
)

type StandardContexts struct {
	CredentialContext
	LoggingContext
	CachingContext
	CachingFileSystem
	CachingPath
	Credentials
}

func (o *StandardContexts) Cache() *tmpcache.Attribute {
	if o.CachingPath.Value != "" {
		return tmpcache.New(o.CachingPath.Value, o.CachingFileSystem.Value)
	}
	if o.CachingContext.Value != nil {
		return tmpcache.Get(o.CachingContext.Value)
	}
	return tmpcache.Get(o.CredentialContext.Value)
}
