package stdopts

import (
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"ocm.software/ocm/api/datacontext/attrs/tmpcache"
	"ocm.software/ocm/api/datacontext/attrs/vfsattr"
	"ocm.software/ocm/api/utils"
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

func (o *StandardContexts) GetCachingFileSystem() vfs.FileSystem {
	if o.CachingFileSystem.Value != nil {
		return o.CachingFileSystem.Value
	}
	if o.CachingContext.Value != nil {
		if fs := o.CachingContext.Value.GetAttributes().GetAttribute(vfsattr.ATTR_KEY); fs != nil {
			return fs.(vfs.FileSystem)
		}
	}
	if o.CredentialContext.Value != nil {
		return utils.FileSystem(vfsattr.Get(o.CredentialContext.Value))
	}
	return osfs.OsFs
}
