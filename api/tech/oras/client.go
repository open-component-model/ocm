package oras

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/containerd/containerd/errdefs"
	ociv1 "github.com/opencontainers/image-spec/specs-go/v1"
	oraserr "oras.land/oras-go/v2/errdef"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
)

type ClientOptions struct {
	Client    *auth.Client
	PlainHTTP bool
}

type Client struct {
	Client    *auth.Client
	PlainHTTP bool
	Ref       string
}

var (
	_ Resolver = &Client{}
	_ Fetcher  = &Client{}
	_ Pusher   = &Client{}
	_ Lister   = &Client{}
)

func New(opts ClientOptions) *Client {
	return &Client{Client: opts.Client, PlainHTTP: opts.PlainHTTP}
}

func (c *Client) Resolve(ctx context.Context, ref string) (string, ociv1.Descriptor, error) {
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

func (c *Client) Fetcher(ctx context.Context, ref string) (Fetcher, error) {
	c.Ref = ref
	return c, nil
}

func (c *Client) Pusher(ctx context.Context, ref string) (Pusher, error) {
	c.Ref = ref
	return c, nil
}

func (c *Client) Lister(ctx context.Context, ref string) (Lister, error) {
	c.Ref = ref
	return c, nil
}

func (c *Client) Push(ctx context.Context, d ociv1.Descriptor, src Source) error {
	reader, err := src.Reader()
	if err != nil {
		return err
	}

	repository, err := c.createRepository(c.Ref)
	if err != nil {
		return err
	}

	if split := strings.Split(c.Ref, ":"); len(split) == 2 {
		// Once we get a reference that contains a tag, we need to re-push that
		// layer with the reference included. PushReference then will tag
		// that layer resulting in the created tag pointing to the right
		// blob data.
		if err := repository.PushReference(ctx, d, reader, c.Ref); err != nil {
			return fmt.Errorf("failed to push tag: %w", err)
		}

		return nil
	}

	// We have a digest, so we push use plain push for the digest.
	// Push here decides if it's a Manifest or a Blob.
	if err := repository.Push(ctx, d, reader); err != nil {
		return fmt.Errorf("failed to push: %w, %s", err, c.Ref)
	}

	return nil
}

func (c *Client) Fetch(ctx context.Context, desc ociv1.Descriptor) (io.ReadCloser, error) {
	src, err := c.createRepository(c.Ref)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve ref %q: %w", c.Ref, err)
	}

	// oras requires a Resolve to happen before a fetch because
	// -1 is an invalid size.
	// manifest is not set in the descriptor
	// We explicitly call resolve on manifest first because it might be
	// that the mediatype is not set at this point.
	rdesc, err := src.Manifests().Resolve(ctx, desc.Digest.String())
	if err != nil {
		if errors.Is(err, oraserr.ErrNotFound) {
			rdesc, err = src.Blobs().Resolve(ctx, desc.Digest.String())
			if err != nil {
				return nil, fmt.Errorf("failed to resolve fetch blob %q: %w", desc.Digest.String(), err)
			}
			delayer := func() (io.ReadCloser, error) {
				return src.Blobs().Fetch(ctx, rdesc)
			}

			return newDelayedReader(delayer)
		}

		return nil, fmt.Errorf("failed to resolve fetch manifest %q: %w", desc.Digest.String(), err)
	}

	// lastly, try a manifest fetch.
	fetch, err := src.Fetch(ctx, rdesc)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch manifest: %w", err)
	}

	return fetch, err
}

func (c *Client) List(ctx context.Context) ([]string, error) {
	var result []string
	src, err := c.createRepository(c.Ref)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve ref %q: %w", c.Ref, err)
	}

	if err := src.Tags(ctx, "", func(tags []string) error {
		result = append(result, tags...)
		return nil
	}); err != nil {
		return nil, fmt.Errorf("failed to list tags: %w", err)
	}

	return result, nil
}

func (c *Client) createRepository(ref string) (*remote.Repository, error) {
	src, err := remote.NewRepository(ref)
	if err != nil {
		return nil, fmt.Errorf("failed to create new repository: %w", err)
	}

	src.Client = c.Client // set up authenticated client.
	src.PlainHTTP = c.PlainHTTP

	return src, nil
}
