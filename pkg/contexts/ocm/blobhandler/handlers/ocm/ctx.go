package ocm

import (
	"github.com/open-component-model/ocm/pkg/blobaccess/blobaccess"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/utils"
)

type BlobSink interface {
	AddBlob(blob blobaccess.BlobAccess) (string, error)
}

// StorageContext is the context information passed for Blobhandler
// registered for context type oci.CONTEXT_TYPE.
type StorageContext interface {
	cpi.StorageContext
	BlobSink
}

type DefaultStorageContext struct {
	cpi.DefaultStorageContext
	Sink    BlobSink
	Payload interface{}
}

func New(repo cpi.Repository, compname string, access BlobSink, impltyp string, payload ...interface{}) StorageContext {
	return &DefaultStorageContext{
		DefaultStorageContext: *cpi.NewDefaultStorageContext(repo, compname, cpi.ImplementationRepositoryType{cpi.CONTEXT_TYPE, impltyp}),
		Sink:                  access,
		Payload:               utils.Optional(payload...),
	}
}

func (c *DefaultStorageContext) AddBlob(blob blobaccess.BlobAccess) (string, error) {
	return c.Sink.AddBlob(blob)
}
