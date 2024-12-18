package oras

import (
	"context"
	"fmt"
	"strings"

	"github.com/containerd/errdefs"
	ociv1 "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2/registry/remote/auth"
)

type OrasPusher struct {
	client *auth.Client
	ref    string
}

func (c *OrasPusher) Push(ctx context.Context, d ociv1.Descriptor, src Source) (retErr error) {
	reader, err := src.Reader()
	if err != nil {
		return err
	}

	repository, err := createRepository(c.ref, c.client, false)
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
