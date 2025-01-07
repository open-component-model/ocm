package oras

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"

	ociv1 "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
)

type OrasFetcher struct {
	client    *auth.Client
	ref       string
	plainHTTP bool
	mu        sync.Mutex
}

func (c *OrasFetcher) Fetch(ctx context.Context, desc ociv1.Descriptor) (io.ReadCloser, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	src, err := createRepository(c.ref, c.client, c.plainHTTP)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve ref %q: %w", c.ref, err)
	}

	// oras requires a Resolve to happen in some cases before a fetch because
	// -1 or 0 are invalid sizes and result in a content-length mismatch error by design.
	// This is a security consideration on ORAS' side.
	// For more information (https://github.com/oras-project/oras-go/issues/822#issuecomment-2325622324)
	//
	// To workaround, we resolve the correct size
	if desc.Size < 1 {
		if desc, err = c.resolveDescriptor(ctx, desc, src); err != nil {
			return nil, err
		}
	}

	// manifest resolve succeeded return the reader directly
	// mediatype of the descriptor should now be set to the correct type.
	reader, err := src.Fetch(ctx, desc)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch blob: %w", err)
	}

	return reader, nil
}

// resolveDescriptor resolves the descriptor by fetching the blob or manifest based on the digest as a reference.
// If the descriptor has a media type, it will be resolved directly.
// If the descriptor has no media type, it will first try to resolve the blob, then the manifest as a fallback.
func (c *OrasFetcher) resolveDescriptor(ctx context.Context, desc ociv1.Descriptor, src *remote.Repository) (ociv1.Descriptor, error) {
	if desc.MediaType != "" {
		var err error
		// if there is a media type, resolve the descriptor directly
		if isManifest(src.ManifestMediaTypes, desc) {
			desc, err = src.Manifests().Resolve(ctx, desc.Digest.String())
		} else {
			desc, err = src.Blobs().Resolve(ctx, desc.Digest.String())
		}
		if err != nil {
			return ociv1.Descriptor{}, fmt.Errorf("failed to resolve descriptor %q: %w", desc.Digest.String(), err)
		}

		return desc, nil
	}

	// if there is no media type, first try the blob, then the manifest
	// To reader: DO NOT fetch manifest first, this can result in high latency calls
	bdesc, err := src.Blobs().Resolve(ctx, desc.Digest.String())
	if err != nil {
		mdesc, merr := src.Manifests().Resolve(ctx, desc.Digest.String())
		if merr != nil {
			// also display the first manifest resolve error
			err = errors.Join(err, merr)

			return ociv1.Descriptor{}, fmt.Errorf("failed to resolve manifest %q: %w", desc.Digest.String(), err)
		}

		return mdesc, nil
	}

	return bdesc, err
}
