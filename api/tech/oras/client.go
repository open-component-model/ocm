package oras

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/containerd/containerd/errdefs"
	"github.com/mandelsoft/logging"
	ociv1 "github.com/opencontainers/image-spec/specs-go/v1"
	oraserr "oras.land/oras-go/v2/errdef"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
)

type ClientOptions struct {
	Client    *auth.Client
	PlainHTTP bool
	Logger    logging.Logger
}

type Client struct {
	client    *auth.Client
	plainHTTP bool
	ref       string
	mu        sync.RWMutex
	logger    logging.Logger
}

var (
	_ Resolver = &Client{}
	_ Fetcher  = &Client{}
	_ Pusher   = &Client{}
	_ Lister   = &Client{}
)

func New(opts ClientOptions) *Client {
	return &Client{client: opts.Client, plainHTTP: opts.PlainHTTP, logger: opts.Logger}
}

func (c *Client) Fetcher(ctx context.Context, ref string) (Fetcher, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.ref = ref
	return c, nil
}

func (c *Client) Pusher(ctx context.Context, ref string) (Pusher, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.ref = ref
	return c, nil
}

func (c *Client) Lister(ctx context.Context, ref string) (Lister, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.ref = ref
	return c, nil
}

func (c *Client) Resolve(ctx context.Context, ref string) (string, ociv1.Descriptor, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	src, err := c.createRepository(ref)
	if err != nil {
		return "", ociv1.Descriptor{}, err
	}

	// We try to first resolve a manifest.
	// _Note_: If there is an error like not found, but we know that the digest exists
	// we can add src.Blobs().Resolve in here. If we do that, note that
	// for Blobs().Resolve `not found` is actually `invalid checksum digest format`.
	// Meaning it will throw that error instead of not found.
	desc, err := src.Resolve(ctx, ref)
	if err != nil {
		if errors.Is(err, oraserr.ErrNotFound) {
			return "", ociv1.Descriptor{}, errdefs.ErrNotFound
		}

		return "", ociv1.Descriptor{}, fmt.Errorf("failed to resolve manifest %q: %w", ref, err)
	}

	return "", desc, nil
}

func (c *Client) Push(ctx context.Context, d ociv1.Descriptor, src Source) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	reader, err := src.Reader()
	if err != nil {
		return err
	}

	repository, err := c.createRepository(c.ref)
	if err != nil {
		return err
	}

	if split := strings.Split(c.ref, ":"); len(split) == 2 {
		// Once we get a reference that contains a tag, we need to re-push that
		// layer with the reference included. PushReference then will tag
		// that layer resulting in the created tag pointing to the right
		// blob data.
		if err := repository.PushReference(ctx, d, reader, c.ref); err != nil {
			return fmt.Errorf("failed to push tag: %w", err)
		}

		return nil
	}

	ok, err := repository.Exists(ctx, d)
	if err != nil {
		return fmt.Errorf("failed to check if repository %q exists: %w", d, err)
	}

	if ok {
		return errdefs.ErrAlreadyExists
	}

	// We have a digest, so we use plain push for the digest.
	// Push here decides if it's a Manifest or a Blob.
	if err := repository.Push(ctx, d, reader); err != nil {
		return fmt.Errorf("failed to push: %w, %s", err, c.ref)
	}

	return nil
}

func (c *Client) Fetch(ctx context.Context, desc ociv1.Descriptor) (io.ReadCloser, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	c.logger.Trace("beginning fetching descriptor", "desc", desc)

	src, err := c.createRepository(c.ref)
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
		c.logger.Trace("description is without size or digest, resolving...", "digest", desc.Digest, "size", desc.Size)
		rdesc, err := src.Manifests().Resolve(ctx, desc.Digest.String())
		if err != nil {
			var berr error
			rdesc, berr = src.Blobs().Resolve(ctx, desc.Digest.String())
			if berr != nil {
				// also display the first manifest resolve error
				err = errors.Join(err, berr)

				return nil, fmt.Errorf("failed to resolve fetch blob %q: %w", desc.Digest.String(), err)
			}

			return src.Blobs().Fetch(ctx, rdesc)
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

	return fetch, err
}

func (c *Client) List(ctx context.Context) ([]string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	src, err := c.createRepository(c.ref)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve ref %q: %w", c.ref, err)
	}

	var result []string
	if err := src.Tags(ctx, "", func(tags []string) error {
		result = append(result, tags...)
		return nil
	}); err != nil {
		return nil, fmt.Errorf("failed to list tags: %w", err)
	}

	return result, nil
}

// createRepository creates a new repository representation using the passed in ref.
// This is a cheap operation.
func (c *Client) createRepository(ref string) (*remote.Repository, error) {
	src, err := remote.NewRepository(ref)
	if err != nil {
		return nil, fmt.Errorf("failed to create new repository: %w", err)
	}

	src.Client = c.client // set up authenticated client.
	src.PlainHTTP = c.plainHTTP

	return src, nil
}
