package oras

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	ociv1 "github.com/opencontainers/image-spec/specs-go/v1"
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

	start := time.Now()
	id := uuid.New()

	log.Printf("START fetch %s; %s: %s\n", id.String(), c.ref, start)
	defer log.Printf("END fetch %s; %s: %s\n", id.String(), c.ref, time.Since(start))

	src, err := createRepository(c.ref, c.client, c.plainHTTP)
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
		log.Printf("Trying to resolve blob %s: %s", id.String(), c.ref)
		bdesc, err := src.Blobs().Resolve(ctx, desc.Digest.String())
		if err != nil {
			log.Printf("Failed to resolve blob, trying to resolve manifest %s: %s", id.String(), c.ref)
			mdesc, merr := src.Manifests().Resolve(ctx, desc.Digest.String())
			if merr != nil {
				// also display the first manifest resolve error
				err = errors.Join(err, merr)

				return nil, fmt.Errorf("failed to resolve manifest %q: %w", desc.Digest.String(), err)
			}

			log.Printf("Fetching manifest %s: %s", id.String(), c.ref)
			fetch, err := src.Fetch(ctx, mdesc)
			if err != nil {
				return nil, fmt.Errorf("failed to fetch manifest: %w", err)
			}

			log.Printf("Manifest fetched; returining %s: %s", id.String(), c.ref)
			return fetch, nil
		}

		// no error
		desc = bdesc
	}

	// manifest resolve succeeded return the reader directly
	// mediatype of the descriptor should now be set to the correct type.
	log.Printf("Blob resolved, fetching reader %s: %s", id.String(), c.ref)
	reader, err := src.Blobs().Fetch(ctx, desc)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch blob: %w", err)
	}
	log.Printf("Reader fetched, returning %s: %s", id.String(), c.ref)

	return reader, nil
}
