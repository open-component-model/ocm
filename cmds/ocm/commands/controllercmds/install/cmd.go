package install

import (
	"context"
	"fmt"
	"time"

	"github.com/fluxcd/pkg/ssa"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"

	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/cmds/ocm/commands/controllercmds/common"
	"ocm.software/ocm/cmds/ocm/commands/controllercmds/names"
	"ocm.software/ocm/cmds/ocm/commands/verbs"
	"ocm.software/ocm/cmds/ocm/common/utils"
)

var (
	Names = names.Controller
	Verb  = verbs.Install
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
	DryRun                   bool
	SkipPreFlightCheck       bool
	InstallPrerequisites     bool
	Silent                   bool
	SM                       *ssa.ResourceManager
}

var _ utils.OCMCommand = (*Command)(nil)

// NewCommand creates a new controller command.
func NewCommand(ctx clictx.Context, names ...string) *cobra.Command {
	return utils.SetupCommand(&Command{BaseCommand: utils.NewBaseCommand(ctx)}, utils.Names(Names, names...)...)
}

func (o *Command) ForName(name string) *cobra.Command {
	return &cobra.Command{
		Use:   "install controller {--version v0.0.1}",
		Short: "Install either a specific or latest version of the ocm-controller. Optionally install prerequisites required by the controller.",
	}
}

func (o *Command) AddFlags(set *pflag.FlagSet) {
	set.StringVarP(&o.Version, "version", "v", "latest", "the version of the controller to install")
	set.StringVarP(&o.BaseURL, "base-url", "u", "https://github.com/open-component-model/ocm-controller/releases", "the base url to the ocm-controller's release page")
	set.StringVarP(&o.ReleaseAPIURL, "release-api-url", "a", "https://api.github.com/repos/open-component-model/ocm-controller/releases", "the base url to the ocm-controller's API release page")
	set.StringVar(&o.CertManagerBaseURL, "cert-manager-base-url", "https://github.com/cert-manager/cert-manager/releases", "the base url to the cert-manager's release page")
	set.StringVar(&o.CertManagerReleaseAPIURL, "cert-manager-release-api-url", "https://api.github.com/repos/cert-manager/cert-manager/releases", "the base url to the cert-manager's API release page")
	set.StringVar(&o.CertManagerVersion, "cert-manager-version", "v1.13.2", "version for cert-manager")
	set.StringVarP(&o.ControllerName, "controller-name", "c", "ocm-controller", "name of the controller that's used for status check")
	set.StringVarP(&o.Namespace, "namespace", "n", "ocm-system", "the namespace into which the controller is installed")
	set.DurationVarP(&o.Timeout, "timeout", "t", 1*time.Minute, "maximum time to wait for deployment to be ready")
	set.BoolVarP(&o.DryRun, "dry-run", "d", false, "if enabled, prints the downloaded manifest file")
	set.BoolVarP(&o.SkipPreFlightCheck, "skip-pre-flight-check", "s", false, "skip the pre-flight check for clusters")
	set.BoolVarP(&o.InstallPrerequisites, "install-prerequisites", "i", true, "install prerequisites required by ocm-controller")
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
	if !o.SkipPreFlightCheck {
		common.Outf(o.Context, o.DryRun, "► running pre-install check\n")
		if err := o.RunPreFlightCheck(ctx); err != nil {
			if o.InstallPrerequisites {
				common.Outf(o.Context, o.DryRun, "► installing prerequisites\n")
				if err := o.installPrerequisites(ctx); err != nil {
					return err
				}

				common.Outf(o.Context, o.DryRun, "✔ successfully installed prerequisites\n")
			} else {
				return fmt.Errorf("✗ failed to run pre-flight check: %w\n", err)
			}
		}
	}

	common.Outf(o.Context, o.DryRun, "► installing ocm-controller with version %s\n", o.Version)
	version := o.Version
	if err := common.Install(
		ctx,
		o.Context,
		sm,
		o.ReleaseAPIURL,
		o.BaseURL,
		"ocm-controller",
		"install.yaml",
		version,
		o.DryRun,
	); err != nil {
		return err
	}

	common.Outf(o.Context, o.DryRun, "✔ ocm-controller successfully installed\n")
	return nil
}

// RunPreFlightCheck checks if the target cluster has the following items:
// - secret containing certificates for the in-cluster registry
// - flux installed.
func (o *Command) RunPreFlightCheck(ctx context.Context) error {
	rcg := genericclioptions.NewConfigFlags(false)
	cfg, err := rcg.ToRESTConfig()
	if err != nil {
		return fmt.Errorf("loading kubeconfig failed: %w", err)
	}

	// bump limits
	cfg.QPS = 100.0
	cfg.Burst = 300

	if err := o.checkCertificateSecretExists(ctx, cfg, rcg); err != nil {
		return fmt.Errorf("ocm-controller requires ocm-registry-tls-certs in ocm-system namespace to exist: %w", err)
	}

	if err := o.checkFluxExists(ctx, cfg, rcg); err != nil {
		return err
	}

	return nil
}

func (o *Command) checkCertificateSecretExists(ctx context.Context, cfg *rest.Config, rcg *genericclioptions.ConfigFlags) error {
	restMapper, err := rcg.ToRESTMapper()
	if err != nil {
		return err
	}

	kubeClient, err := client.New(cfg, client.Options{Mapper: restMapper, Scheme: newScheme()})
	if err != nil {
		return err
	}

	s := &corev1.Secret{}
	return kubeClient.Get(ctx, types.NamespacedName{
		Name:      "ocm-registry-tls-certs",
		Namespace: "ocm-system",
	}, s)
}

func (o *Command) checkFluxExists(ctx context.Context, cfg *rest.Config, rcg *genericclioptions.ConfigFlags) error {
	restMapper, err := rcg.ToRESTMapper()
	if err != nil {
		return err
	}

	kubeClient, err := client.New(cfg, client.Options{Mapper: restMapper, Scheme: newScheme()})
	if err != nil {
		return err
	}

	s := &corev1.Namespace{}
	return kubeClient.Get(ctx, types.NamespacedName{
		Name: "flux-system",
	}, s)
}
