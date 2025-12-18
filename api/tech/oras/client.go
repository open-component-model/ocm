package oras

import (
	"context"
	"errors"
	"fmt"

	"github.com/containerd/containerd/errdefs"
	"github.com/mandelsoft/logging"
	"github.com/moby/locker"
	ociv1 "github.com/opencontainers/image-spec/specs-go/v1"
	oraserr "oras.land/oras-go/v2/errdef"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
)

type ClientOptions struct {
	Client    *auth.Client
	PlainHTTP bool
	Logger    logging.Logger
	Lock      *locker.Locker
}

type Client struct {
	client    *auth.Client
	plainHTTP bool
	logger    logging.Logger
}

var _ Resolver = &Client{}

func New(opts ClientOptions) *Client {
	return &Client{client: opts.Client, plainHTTP: opts.PlainHTTP, logger: opts.Logger}
}

func (c *Client) Fetcher(ctx context.Context, ref string) (Fetcher, error) {
	return &OrasFetcher{client: c.client, ref: ref, plainHTTP: c.plainHTTP}, nil
}

func (c *Client) Pusher(ctx context.Context, ref string) (Pusher, error) {
	return &OrasPusher{client: c.client, ref: ref, plainHTTP: c.plainHTTP}, nil
}

func (c *Client) Lister(ctx context.Context, ref string) (Lister, error) {
	return &OrasLister{client: c.client, ref: ref, plainHTTP: c.plainHTTP}, nil
}

func (c *Client) Resolve(ctx context.Context, ref string) (string, ociv1.Descriptor, error) {
	src, err := createRepository(ref, c.client, c.plainHTTP)
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

// createRepository creates a new repository representation using the passed in ref.
// This is a cheap operation.
func createRepository(ref string, client *auth.Client, plain bool) (*remote.Repository, error) {
	src, err := remote.NewRepository(ref)
	if err != nil {
		return nil, fmt.Errorf("failed to create new repository: %w", err)
	}

	src.Client = client // set up authenticated client.
	src.PlainHTTP = plain

	return src, nil
}
