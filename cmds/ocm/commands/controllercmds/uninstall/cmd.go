package uninstall

import (
	"context"
	"fmt"
	"time"

	"github.com/fluxcd/pkg/ssa"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/api/utils/out"
	"ocm.software/ocm/cmds/ocm/commands/controllercmds/common"
	"ocm.software/ocm/cmds/ocm/commands/controllercmds/names"
	"ocm.software/ocm/cmds/ocm/commands/verbs"
	"ocm.software/ocm/cmds/ocm/common/utils"
)

var (
	Names = names.Controller
	Verb  = verbs.Uninstall
)

type Command struct {
	utils.BaseCommand
	Namespace                string
	ControllerName           string
	Timeout                  time.Duration
	Version                  string
	BaseURL                  string
	ReleaseAPIURL            string
	CertManagerBaseURL       string
	CertManagerReleaseAPIURL string
	CertManagerVersion       string
	SM                       *ssa.ResourceManager
	UninstallPrerequisites   bool
	Silent                   bool
	DryRun                   bool
}

var _ utils.OCMCommand = (*Command)(nil)

// NewCommand creates a new controller command.
func NewCommand(ctx clictx.Context, names ...string) *cobra.Command {
	return utils.SetupCommand(&Command{BaseCommand: utils.NewBaseCommand(ctx)}, utils.Names(Names, names...)...)
}

func (o *Command) ForName(name string) *cobra.Command {
	return &cobra.Command{
		Use:   "uninstall controller",
		Short: "Uninstalls the ocm-controller and all of its dependencies",
	}
}

// AddFlags for the known item to delete.
func (o *Command) AddFlags(set *pflag.FlagSet) {
	set.StringVarP(&o.Version, "version", "v", "latest", "the version of the controller to install")
	set.StringVarP(&o.BaseURL, "base-url", "u", "https://github.com/ocm-controller/releases", "the base url to the ocm-controller's release page")
	set.StringVarP(&o.ReleaseAPIURL, "release-api-url", "a", "https://api.github.com/repos/open-component-model/ocm-controller/releases", "the base url to the ocm-controller's API release page")
	set.StringVar(&o.CertManagerBaseURL, "cert-manager-base-url", "https://github.com/cert-manager/cert-manager/releases", "the base url to the cert-manager's release page")
	set.StringVar(&o.CertManagerReleaseAPIURL, "cert-manager-release-api-url", "https://api.github.com/repos/cert-manager/cert-manager/releases", "the base url to the cert-manager's API release page")
	set.StringVar(&o.CertManagerVersion, "cert-manager-version", "v1.13.2", "version for cert-manager")
	set.StringVarP(&o.ControllerName, "controller-name", "c", "ocm-controller", "name of the controller that's used for status check")
	set.StringVarP(&o.Namespace, "namespace", "n", "ocm-system", "the namespace into which the controller is installed")
	set.DurationVarP(&o.Timeout, "timeout", "t", 1*time.Minute, "maximum time to wait for deployment to be ready")
	set.BoolVarP(&o.UninstallPrerequisites, "uninstall-prerequisites", "p", false, "uninstall prerequisites required by ocm-controller")
	set.BoolVarP(&o.DryRun, "dry-run", "d", false, "if enabled, prints the downloaded manifest file")
	set.BoolVarP(&o.Silent, "silent", "l", false, "don't fail on error")
}

func (o *Command) Complete(args []string) error {
	return nil
}

func (o *Command) Run() (err error) {
	defer func() {
		// don't return any errors
		if o.Silent {
			err = nil
		}
	}()

	kubeconfigArgs := genericclioptions.NewConfigFlags(false)
	sm, err := NewResourceManager(kubeconfigArgs)
	if err != nil {
		return fmt.Errorf("✗ failed to create resource manager: %w", err)
	}

	o.SM = sm
	ctx := context.Background()

	out.Outf(o.Context, "► uninstalling ocm-controller with version %s\n", o.Version)
	if err := common.Uninstall(
		ctx,
		o.Context,
		sm,
		o.ReleaseAPIURL,
		o.BaseURL,
		"ocm-controller",
		"install.yaml",
		o.Version,
		o.DryRun,
	); err != nil {
		return err
	}

	out.Outf(o.Context, "✔ ocm-controller successfully uninstalled\n")

	if o.UninstallPrerequisites {
		out.Outf(o.Context, "► uninstalling cert-manager and issuers\n")
		if err := o.uninstallPrerequisites(ctx); err != nil {
			return fmt.Errorf("✗ failed to uninstall pre-requesits: %w\n", err)
		}

		out.Outf(o.Context, "✔ successfully uninstalled prerequisites\n")
	}

	return nil
}
