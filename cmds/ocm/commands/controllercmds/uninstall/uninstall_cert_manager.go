package uninstall

import (
	"context"
	"fmt"
	"os"

	_ "embed"

	"github.com/fluxcd/pkg/ssa"
	"github.com/mandelsoft/filepath/pkg/filepath"

	"ocm.software/ocm/api/utils/out"
	"ocm.software/ocm/cmds/ocm/commands/controllercmds/common"
)

//go:embed issuer/registry_certificate.yaml
var issuer []byte

func (o *Command) uninstallPrerequisites(ctx context.Context) error {
	out.Outf(o.Context, "► uninstalling cert-manager with version %s\n", o.CertManagerVersion)

	if err := o.removeRegistryCertificate(); err != nil {
		return fmt.Errorf("✗ failed to create registry certificate: %w", err)
	}

	version := o.CertManagerVersion
	if err := common.Uninstall(
		ctx,
		o.Context,
		o.SM,
		o.CertManagerReleaseAPIURL,
		o.CertManagerBaseURL,
		"cert-manager",
		"cert-manager.yaml",
		version,
		o.DryRun,
	); err != nil {
		return err
	}

	out.Outf(o.Context, "✔ cert-manager successfully uninstalled\n")

	return nil
}

func (o *Command) removeRegistryCertificate() error {
	out.Outf(o.Context, "► remove certificate for internal registry\n")
	temp, err := os.MkdirTemp("", "issuer")
	if err != nil {
		return fmt.Errorf("failed to create temp folder: %w", err)
	}
	defer os.RemoveAll(temp)

	path := filepath.Join(temp, "issuer.yaml")
	if err := os.WriteFile(path, issuer, 0o600); err != nil {
		return fmt.Errorf("failed to write issuer.yaml file at location: %w", err)
	}

	objects, err := common.ReadObjects(path)
	if err != nil {
		return fmt.Errorf("failed to construct objects to apply: %w", err)
	}

	if _, err := o.SM.DeleteAll(context.Background(), objects, ssa.DefaultDeleteOptions()); err != nil {
		return fmt.Errorf("failed to delete manifests: %w", err)
	}

	return nil
}
