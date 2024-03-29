package install

import (
	"context"
	"fmt"
	"os"

	_ "embed"

	"github.com/fluxcd/pkg/ssa"
	"github.com/mandelsoft/filepath/pkg/filepath"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/open-component-model/ocm/pkg/out"
)

//go:embed issuer/registry_certificate.yaml
var issuer []byte

func (o *Command) installPrerequisites(ctx context.Context) error {
	out.Outf(o.Context, "► installing cert-manager with version %s\n", o.CertManagerVersion)

	version := o.CertManagerVersion
	if err := o.installManifest(
		ctx,
		o.CertManagerReleaseAPIURL,
		o.CertManagerBaseURL,
		"cert-manager",
		"cert-manager.yaml",
		version,
	); err != nil {
		return err
	}

	out.Outf(o.Context, "✔ cert-manager successfully installed\n")
	out.Outf(o.Context, "► creating certificate for internal registry\n")

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

	kubeconfigArgs := genericclioptions.NewConfigFlags(false)
	sm, err := NewResourceManager(kubeconfigArgs)
	if err != nil {
		return fmt.Errorf("failed to create resource manager: %w", err)
	}

	objects, err := readObjects(path)
	if err != nil {
		return fmt.Errorf("failed to construct objects to apply: %w", err)
	}

	if _, err := sm.ApplyAllStaged(context.Background(), objects, ssa.DefaultApplyOptions()); err != nil {
		return fmt.Errorf("failed to apply manifests: %w", err)
	}

	return nil
}
