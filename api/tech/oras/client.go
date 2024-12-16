package oras

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/containerd/containerd/content"
	"github.com/containerd/containerd/errdefs"
	"github.com/opencontainers/go-digest"
	ociv1 "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"

	"ocm.software/ocm/api/tech/regclient"
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

type pushRequest struct{}

func (p *pushRequest) Commit(ctx context.Context, size int64, expected digest.Digest, opts ...content.Opt) error {
	return nil
}

func (p *pushRequest) Status() (content.Status, error) {
	return content.Status{}, nil
}

var _ regclient.PushRequest = &pushRequest{}

var (
	_ regclient.Resolver = &Client{}
	_ regclient.Fetcher  = &Client{}
	_ regclient.Pusher   = &Client{}
	_ regclient.Lister   = &Client{}
)

func New(opts ClientOptions) *Client {
	return &Client{Client: opts.Client, PlainHTTP: opts.PlainHTTP}
}

func (c *Client) Resolve(ctx context.Context, ref string) (string, ociv1.Descriptor, error) {
	src, err := c.resolveRef(ref)
	if err != nil {
		return "", ociv1.Descriptor{}, err
	}

	// We try to first resolve a manifest.
	desc, err := src.Resolve(ctx, ref)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			// Then we use the blob store to resolve.
			desc, err := src.Blobs().Resolve(ctx, ref)
			if err != nil {
				if strings.Contains(err.Error(), "not found") {
					return "", ociv1.Descriptor{}, errdefs.ErrNotFound
				}

				return "", ociv1.Descriptor{}, fmt.Errorf("failed to resolve blob: %w", err)
			}

			return ref, desc, nil
		}

		return "", ociv1.Descriptor{}, fmt.Errorf("failed to resolve manifest %q: %w", ref, err)
	}

	return "", desc, nil
}

func (c *Client) Fetcher(ctx context.Context, ref string) (regclient.Fetcher, error) {
	c.Ref = ref
	return c, nil
}

func (c *Client) Pusher(ctx context.Context, ref string) (regclient.Pusher, error) {
	c.Ref = ref
	return c, nil
}

func (c *Client) Lister(ctx context.Context, ref string) (regclient.Lister, error) {
	c.Ref = ref
	return c, nil
}

func (c *Client) Push(ctx context.Context, d ociv1.Descriptor, src regclient.Source) (regclient.PushRequest, error) {
	reader, err := src.Reader()
	if err != nil {
		return nil, err
	}

	repository, err := c.resolveRef(c.Ref)
	if err != nil {
		return nil, err
	}

	if split := strings.Split(c.Ref, ":"); len(split) == 2 {
		// Once we get a reference that contains a tag, we need to re-push that
		// layer with the reference included. PushReference pushes a blob or a
		// manifest.
		if err := repository.PushReference(ctx, d, reader, c.Ref); err != nil {
			return nil, fmt.Errorf("failed to push tag: %w", err)
		}

		return &pushRequest{}, nil
	}

	// We have a digest, so we push use plain push for the digest.
	// Push here decides if it's a Manifest or a Blob.
	if err := repository.Push(ctx, d, reader); err != nil {
		return nil, fmt.Errorf("failed to push: %w, %s", err, c.Ref)
	}

	return &pushRequest{}, nil
}

func (c *Client) Fetch(ctx context.Context, desc ociv1.Descriptor) (io.ReadCloser, error) {
	src, err := c.resolveRef(c.Ref)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve ref %q: %w", c.Ref, err)
	}

	// manifest is not set in the descriptor
	// src.Resolve is a manifest().resolve
	rdesc, err := src.Resolve(ctx, desc.Digest.String())
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
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

	fetch, err := src.Fetch(ctx, rdesc)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch manifest: %w", err)
	}

	return fetch, err
}

func (c *Client) List(ctx context.Context) ([]string, error) {
	var result []string
	src, err := c.resolveRef(c.Ref)
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

func (c *Client) resolveRef(ref string) (*remote.Repository, error) {
	src, err := remote.NewRepository(ref)
	if err != nil {
		return nil, fmt.Errorf("failed to create new repository: %w", err)
	}

	src.Client = c.Client // set up authenticated client.
	src.PlainHTTP = c.PlainHTTP

	return src, nil
}
