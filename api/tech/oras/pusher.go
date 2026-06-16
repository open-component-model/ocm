package oras

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/containerd/errdefs"

	"oras.land/oras-go/v2/errdef"
	"oras.land/oras-go/v2/registry"
	"oras.land/oras-go/v2/registry/remote/auth"

	ociv1 "github.com/opencontainers/image-spec/specs-go/v1"

	"ocm.software/ocm/api/oci/ociutils"
	"ocm.software/ocm/api/utils/logging"
)

// PushExistsCheckEnvVar gates the optional HEAD pre-check performed before
// pushing a blob to an OCI registry. When set to a truthy value (parsed via
// strconv.ParseBool, e.g. "1", "true"), OrasPusher.Push calls
// repository.Exists() first and short-circuits with errdefs.ErrAlreadyExists
// if the blob is already present.
//
// The check is opt-in because it adds an extra round-trip per blob and
// serialises Exists/Push pairs in the hot path of concurrent transfers
// (see PR #1676). Enable it when a registry's Push response is unreliable
// and the redundant HEAD is preferable to a failed upload.
const PushExistsCheckEnvVar = "OCM_OCI_PUSH_EXISTS_CHECK"

// pushExistsCheckEnabled reports whether the optional pre-push Exists() check
// is requested via PushExistsCheckEnvVar. The env var is consulted exactly
// once at package initialisation; unset, empty, or unparsable values disable
// the check.
//
// strconv.ParseBool returns (false, err) for "" - the os.Getenv result for an
// unset variable — so a single call covers the unset and unparsable cases.
var pushExistsCheckEnabled = func() bool {
	enabled, _ := strconv.ParseBool(os.Getenv(PushExistsCheckEnvVar))
	logging.Logger().Debug("oras pre-push Exists() check gate resolved",
		"envVar", PushExistsCheckEnvVar, "enabled", enabled)
	return enabled
}()

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
	defer func() {
		reader.Close()
	}()

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
			if errors.Is(err, errdef.ErrAlreadyExists) {
				return errdefs.ErrAlreadyExists
			}
			return fmt.Errorf("failed to push tag: %w, %s", err, c.ref)
		}

		return nil
	}

	if pushExistsCheckEnabled {
		ok, err := repository.Exists(ctx, d)
		if err != nil {
			return fmt.Errorf("failed to check if repository %q exists: %w", ref.Repository, err)
		}
		if ok {
			logging.Logger().Debug("pre-push Exists() check short-circuited; blob already present",
				"ref", c.ref, "digest", d.Digest.String())
			return errdefs.ErrAlreadyExists
		}
	}

	if err := repository.Push(ctx, d, reader); err != nil {
		if errors.Is(err, errdef.ErrAlreadyExists) {
			return errdefs.ErrAlreadyExists
		}
		return fmt.Errorf("failed to push: %w, %s", err, c.ref)
	}

	return nil
}
