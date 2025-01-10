package comparch

import (
	"fmt"

	"github.com/mandelsoft/goutils/errors"
	"github.com/opencontainers/go-digest"

	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/localblob"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/localfsblob"
	"ocm.software/ocm/api/ocm/extensions/attrs/compatattr"
	storagecontext "ocm.software/ocm/api/ocm/extensions/blobhandler/handlers/ocm"
	"ocm.software/ocm/api/ocm/extensions/repositories/comparch"
	common "ocm.software/ocm/api/utils/misc"
)

// Deprecated: Component Archive (CA) - https://kubernetes.slack.com/archives/C05UWBE8R1D/p1734357630853489
func init() {
	cpi.RegisterBlobHandler(NewBlobHandler(), cpi.ForRepo(cpi.CONTEXT_TYPE, comparch.Type))
}

////////////////////////////////////////////////////////////////////////////////

// Deprecated: Component Archive (CA) - https://kubernetes.slack.com/archives/C05UWBE8R1D/p1734357630853489
// blobHandler is the default handling to store local blobs as local blobs.
type blobHandler struct{}

// Deprecated: Component Archive (CA) - https://kubernetes.slack.com/archives/C05UWBE8R1D/p1734357630853489
func NewBlobHandler() cpi.BlobHandler {
	return &blobHandler{}
}

// Deprecated: Component Archive (CA) - https://kubernetes.slack.com/archives/C05UWBE8R1D/p1734357630853489
func (b *blobHandler) StoreBlob(blob cpi.BlobAccess, artType, hint string, global cpi.AccessSpec, ctx cpi.StorageContext) (cpi.AccessSpec, error) {
	ocmctx, ok := ctx.(storagecontext.StorageContext)
	if !ok {
		return nil, fmt.Errorf("failed to assert type %T to storagecontext.StorageContext", ctx)
	}

	if blob == nil {
		return nil, errors.New("a resource has to be defined")
	}
	ref, err := ocmctx.AddBlob(blob)
	if err != nil {
		return nil, err
	}
	path := common.DigestToFileName(digest.Digest(ref))
	if compatattr.Get(ctx.GetContext()) {
		return localfsblob.New(path, blob.MimeType()), nil
	} else {
		return localblob.New(path, hint, blob.MimeType(), global), nil
	}
}
