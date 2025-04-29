package ocm

import (
	"github.com/mandelsoft/goutils/general"

	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/utils/blobaccess/blobaccess"
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
		Payload:               general.Optional(payload...),
	}
}

func (c *DefaultStorageContext) AddBlob(blob blobaccess.BlobAccess) (string, error) {
	return c.Sink.AddBlob(blob)
}
