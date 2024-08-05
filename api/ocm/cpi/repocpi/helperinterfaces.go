package repocpi

import (
	"fmt"

	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/utils/refmgmt/resource"
)

var (
	ErrClosed      = resource.ErrClosed
	ErrTempVersion = fmt.Errorf("temporary component version cannot be updated")
)

// BlobContainer is the interface for an element capable to store blobs.
type BlobContainer interface {
	GetBlob(name string) (cpi.DataAccess, error)

	// GetStorageContext creates a storage context for blobs
	// that is used to feed blob handlers for specific blob storage methods.
	// If no handler accepts the blob, the AddBlobFor method will
	// be used to store the blob
	GetStorageContext() cpi.StorageContext

	// AddBlob stores a local blob together with the component and
	// potentially provides a global reference according to the OCI distribution spec
	// if the blob described an oci artifact.
	// The resulting access information (global and local) is provided as
	// an access method specification usable in a component descriptor.
	// This is the direct technical storage, without caring about any handler.
	AddBlob(blob cpi.BlobAccess, refName string, global cpi.AccessSpec) (cpi.AccessSpec, error)
}
