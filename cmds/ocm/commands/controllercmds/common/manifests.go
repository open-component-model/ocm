package common

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/fluxcd/pkg/ssa"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/api/utils/out"
)

func Install(ctx context.Context, octx clictx.Context, sm *ssa.ResourceManager, releaseURL, baseURL, manifest, filename, version string, dryRun bool) error {
	objects, err := fetchObjects(ctx, octx, releaseURL, baseURL, manifest, filename, version, dryRun)
	if err != nil {
		return fmt.Errorf("✗ failed to construct objects to apply: %w", err)
	}

	// dry run was set to true, no objects are returned
	if len(objects) == 0 {
		return nil
	}

	if _, err := sm.ApplyAllStaged(context.Background(), objects, ssa.DefaultApplyOptions()); err != nil {
		return fmt.Errorf("✗ failed to apply manifests: %w", err)
	}

	Outf(octx, dryRun, "► waiting for ocm deployment to be ready\n")
	if err = sm.Wait(objects, ssa.DefaultWaitOptions()); err != nil {
		return fmt.Errorf("✗ failed to wait for objects to be ready: %w", err)
	}

	return nil
}

func Uninstall(ctx context.Context, octx clictx.Context, sm *ssa.ResourceManager, releaseURL, baseURL, manifest, filename, version string, dryRun bool) error {
	objects, err := fetchObjects(ctx, octx, releaseURL, baseURL, manifest, filename, version, dryRun)
	if err != nil {
		return fmt.Errorf("✗ failed to construct objects to apply: %w", err)
	}

	// dry run was set to true, no objects are returned
	if len(objects) == 0 {
		return nil
	}

	if _, err := sm.DeleteAll(context.Background(), objects, ssa.DefaultDeleteOptions()); err != nil {
		return fmt.Errorf("✗ failed to delete manifests: %w", err)
	}

	Outf(octx, dryRun, "► waiting for ocm deployment to be deleted\n")
	if err = sm.WaitForTermination(objects, ssa.DefaultWaitOptions()); err != nil {
		return fmt.Errorf("✗ failed to wait for objects to be deleted: %w", err)
	}

	return nil
}

func fetchObjects(ctx context.Context, octx clictx.Context, releaseURL, baseURL, manifest, filename, version string, dryRun bool) ([]*unstructured.Unstructured, error) {
	if version == "latest" {
		latest, err := getLatestVersion(ctx, releaseURL)
		if err != nil {
			return nil, fmt.Errorf("✗ failed to retrieve latest version for %s: %w", manifest, err)
		}
		Outf(octx, dryRun, "► got latest version %q\n", latest)
		version = latest
	} else {
		exists, err := existingVersion(ctx, releaseURL, version)
		if err != nil {
			return nil, fmt.Errorf("✗ failed to check if version exists: %w", err)
		}
		if !exists {
			return nil, fmt.Errorf("✗ version %q does not exist", version)
		}
	}

	temp, err := os.MkdirTemp("", manifest+"-download")
	if err != nil {
		return nil, fmt.Errorf("✗ failed to create temp folder: %w", err)
	}
	defer os.RemoveAll(temp)

	if err := fetch(ctx, baseURL, version, temp, filename); err != nil {
		return nil, fmt.Errorf("✗ failed to download install.yaml file: %w", err)
	}

	path := filepath.Join(temp, filename)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("✗ failed to find %s file at location: %w", filename, err)
	}
	Outf(octx, dryRun, "✔ successfully fetched install file\n")
	if dryRun {
		content, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("✗ failed to read %s file at location: %w", filename, err)
		}

		out.Out(octx, string(content))

		return nil, nil
	}
	Outf(octx, dryRun, "► applying to cluster...\n")

	objects, err := ReadObjects(path)
	if err != nil {
		return nil, fmt.Errorf("✗ failed to construct objects to apply: %w", err)
	}

	return objects, nil
}
