package oras

import (
	"context"
	"fmt"

	"github.com/containerd/errdefs"
	ociv1 "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2/registry"
	"oras.land/oras-go/v2/registry/remote/auth"

	"ocm.software/ocm/api/oci/ociutils"
)

type OrasPusher struct {
	client    *auth.Client
	ref       string
	plainHTTP bool
}

func (c *OrasPusher) Push(ctx context.Context, d ociv1.Descriptor, src Source) (retErr error) {
	reader, err := src.Reader()
	if err != nil {
		return err
	}

	repository, err := createRepository(c.ref, c.client, c.plainHTTP)
	if err != nil {
		return err
	}

	ref, err := registry.ParseReference(c.ref)
	if err != nil {
		return fmt.Errorf("failed to parse reference %q: %w", c.ref, err)
	}

	vers, err := ociutils.ParseVersion(ref.Reference)
	if err != nil {
		return fmt.Errorf("failed to parse version %q: %w", ref.Reference, err)
	}

	if vers.IsTagged() {
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
		return fmt.Errorf("failed to check if repository %q exists: %w", ref.Repository, err)
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
