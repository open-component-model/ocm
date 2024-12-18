package oras

import (
	"context"
	"errors"
	"fmt"
	"io"

	ociv1 "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2/registry/remote/auth"
)

type OrasFetcher struct {
	client *auth.Client
	ref    string
}

func (c *OrasFetcher) Fetch(ctx context.Context, desc ociv1.Descriptor) (io.ReadCloser, error) {
	src, err := createRepository(c.ref, c.client, false)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve ref %q: %w", c.ref, err)
	}

	// oras requires a Resolve to happen before a fetch because
	// -1 or 0 are invalid sizes and result in a content-length mismatch error by design.
	// This is a security consideration on ORAS' side.
	// For more information (https://github.com/oras-project/oras-go/issues/822#issuecomment-2325622324)
	// We explicitly call resolve on manifest first because it might be
	// that the mediatype is not set at this point so we don't want ORAS to try to
	// select the wrong layer to fetch from.
	if desc.Size < 1 || desc.Digest == "" {
		rdesc, err := src.Manifests().Resolve(ctx, desc.Digest.String())
		if err != nil {
			var berr error
			rdesc, berr = src.Blobs().Resolve(ctx, desc.Digest.String())
			if berr != nil {
				// also display the first manifest resolve error
				err = errors.Join(err, berr)

				return nil, fmt.Errorf("failed to resolve fetch blob %q: %w", desc.Digest.String(), err)
			}

			reader, err := src.Blobs().Fetch(ctx, rdesc)
			if err != nil {
				return nil, fmt.Errorf("failed to fetch blob: %w", err)
			}

			return reader, nil
		}

		// no error
		desc = rdesc
	}

	// manifest resolve succeeded return the reader directly
	// mediatype of the descriptor should now be set to the correct type.
	fetch, err := src.Fetch(ctx, desc)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch manifest: %w", err)
	}

	return fetch, nil
}
