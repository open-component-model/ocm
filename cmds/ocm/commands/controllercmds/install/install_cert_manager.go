package install

import (
	"context"
	_ "embed"
	"fmt"
	"os"

	"github.com/fluxcd/pkg/ssa"
	"github.com/mandelsoft/filepath/pkg/filepath"
	"ocm.software/ocm/cmds/ocm/commands/controllercmds/common"
)

//go:embed issuer/registry_certificate.yaml
var issuer []byte

func (o *Command) installPrerequisites(ctx context.Context) error {
	common.Outf(o.Context, o.DryRun, "► installing cert-manager with version %s\n", o.CertManagerVersion)

	if err := common.Install(ctx, o.Context, o.SM, o.CertManagerReleaseAPIURL, o.CertManagerBaseURL, "cert-manager", "cert-manager.yaml", o.CertManagerVersion, o.DryRun); err != nil {
		return err
	}

	common.Outf(o.Context, o.DryRun, "✔ cert-manager successfully installed\n")

	if o.DryRun {
		return nil
	}

	common.Outf(o.Context, o.DryRun, "► creating certificate for internal registry\n")

	if err := o.createRegistryCertificate(); err != nil {
		return fmt.Errorf("✗ failed to create registry certificate: %w", err)
	}

	return nil
}

func (o *Command) createRegistryCertificate() error {
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

	if _, err := o.SM.ApplyAllStaged(context.Background(), objects, ssa.DefaultApplyOptions()); err != nil {
		return fmt.Errorf("failed to apply manifests: %w", err)
	}

	return nil
}
