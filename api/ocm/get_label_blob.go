package ocm

import (
	"fmt"

	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
)

// GetBlobValue returns the blob value for a Label.
func GetBlobValue(in *metav1.Label, acc ComponentVersionAccess) ([]byte, error) {
	ctx := acc.GetContext()
	blobAccess, err := ctx.AccessSpecForSpec(in.Access)
	if err != nil {
		return nil, fmt.Errorf("failed to construct access spec for label %s: %w", in.Name, err)
	}

	// TODO: Q: are we supporting none local blobs?
	if !blobAccess.IsLocal(ctx) {
		return nil, fmt.Errorf("label blob %s is not local", in.Name)
	}

	method, err := blobAccess.AccessMethod(acc)
	if err != nil {
		return nil, fmt.Errorf("failed to construct access method for label %s: %w", in.Name, err)
	}

	blob, err := method.Get()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve blob content for label %s: %w", in.Name, err)
	}

	return blob, nil
}
