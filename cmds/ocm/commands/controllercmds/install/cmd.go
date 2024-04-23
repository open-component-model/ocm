// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package install

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/fluxcd/pkg/ssa"
	"github.com/mandelsoft/filepath/pkg/filepath"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/open-component-model/ocm/cmds/ocm/commands/controllercmds/names"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
	"github.com/open-component-model/ocm/pkg/out"
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
	SM                       *ssa.ResourceManager
}

var _ utils.OCMCommand = (*Command)(nil)

// NewCommand creates a new controller cdommand.
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
}

func (o *Command) Complete(args []string) error {
	return nil
}

func (o *Command) Run() error {
	kubeconfigArgs := genericclioptions.NewConfigFlags(false)
	sm, err := NewResourceManager(kubeconfigArgs)
	if err != nil {
		return fmt.Errorf("✗ failed to create resource manager: %w", err)
	}

	o.SM = sm

	ctx := context.Background()
	if !o.SkipPreFlightCheck {
		out.Outf(o.Context, "► running pre-install check\n")
		if err := o.RunPreFlightCheck(ctx); err != nil {
			if o.InstallPrerequisites {
				out.Outf(o.Context, "► installing prerequisites\n")
				if err := o.installPrerequisites(ctx); err != nil {
					return err
				}

				out.Outf(o.Context, "✔ successfully installed prerequisites\n")
			} else {
				return fmt.Errorf("✗ failed to run pre-flight check: %w\n", err)
			}
		}
	}

	out.Outf(o.Context, "► installing ocm-controller with version %s\n", o.Version)
	version := o.Version
	if err := o.installManifest(
		ctx,
		o.ReleaseAPIURL,
		o.BaseURL,
		"ocm-controller",
		"install.yaml",
		version,
	); err != nil {
		return err
	}

	out.Outf(o.Context, "✔ ocm-controller successfully installed\n")
	return nil
}

func (o *Command) installManifest(ctx context.Context, releaseURL, baseURL, manifest, filename, version string) error {
	if version == "latest" {
		latest, err := o.getLatestVersion(ctx, releaseURL)
		if err != nil {
			return fmt.Errorf("✗ failed to retrieve latest version for %s: %w", manifest, err)
		}
		out.Outf(o.Context, "► got latest version %q\n", latest)
		version = latest
	} else {
		exists, err := o.existingVersion(ctx, releaseURL, version)
		if err != nil {
			return fmt.Errorf("✗ failed to check if version exists: %w", err)
		}
		if !exists {
			return fmt.Errorf("✗ version %q does not exist", version)
		}
	}

	temp, err := os.MkdirTemp("", manifest+"-download")
	if err != nil {
		return fmt.Errorf("✗ failed to create temp folder: %w", err)
	}
	defer os.RemoveAll(temp)

	if err := o.fetch(ctx, baseURL, version, temp, filename); err != nil {
		return fmt.Errorf("✗ failed to download install.yaml file: %w", err)
	}

	path := filepath.Join(temp, filename)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("✗ failed to find %s file at location: %w", filename, err)
	}
	out.Outf(o.Context, "✔ successfully fetched install file\n")
	if o.DryRun {
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("✗ failed to read %s file at location: %w", filename, err)
		}
		out.Outf(o.Context, string(content))
		return nil
	}
	out.Outf(o.Context, "► applying to cluster...\n")

	objects, err := readObjects(path)
	if err != nil {
		return fmt.Errorf("✗ failed to construct objects to apply: %w", err)
	}

	if _, err := o.SM.ApplyAllStaged(context.Background(), objects, ssa.DefaultApplyOptions()); err != nil {
		return fmt.Errorf("✗ failed to apply manifests: %w", err)
	}

	out.Outf(o.Context, "► waiting for ocm deployment to be ready\n")
	if err = o.SM.Wait(objects, ssa.DefaultWaitOptions()); err != nil {
		return fmt.Errorf("✗ failed to wait for objects to be ready: %w", err)
	}

	return nil
}

// getLatestVersion calls the GitHub API and returns the latest released version.
func (o *Command) getLatestVersion(ctx context.Context, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url+"/latest", nil)
	if err != nil {
		return "", err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("GitHub API call failed: %w", err)
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	type meta struct {
		Tag string `json:"tag_name"`
	}
	var m meta
	if err := json.NewDecoder(res.Body).Decode(&m); err != nil {
		return "", fmt.Errorf("decoding GitHub API response failed: %w", err)
	}

	return m.Tag, err
}

// existingVersion calls the GitHub API to confirm the given version does exist.
func (o *Command) existingVersion(ctx context.Context, url, version string) (bool, error) {
	if !strings.HasPrefix(version, "v") {
		version = "v" + version
	}

	ghURL := fmt.Sprintf(url+"/tags/%s", version)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ghURL, nil)
	if err != nil {
		return false, fmt.Errorf("GitHub API call failed: %w", err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, err
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	switch res.StatusCode {
	case http.StatusOK:
		return true, nil
	case http.StatusNotFound:
		return false, nil
	default:
		return false, fmt.Errorf("GitHub API returned an unexpected status code (%d)", res.StatusCode)
	}
}

func (o *Command) fetch(ctx context.Context, url, version, dir, filename string) error {
	ghURL := fmt.Sprintf("%s/latest/download/%s", url, filename)
	if strings.HasPrefix(version, "v") {
		ghURL = fmt.Sprintf("%s/download/%s/%s", url, version, filename)
	}

	req, err := http.NewRequest(http.MethodGet, ghURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create HTTP request for %s, error: %w", ghURL, err)
	}

	resp, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("failed to download manifests.tar.gz from %s, error: %w", ghURL, err)
	}
	defer resp.Body.Close()

	// check response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download %s from %s, status: %s", filename, ghURL, resp.Status)
	}

	wf, err := os.OpenFile(filepath.Join(dir, filename), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o777)
	if err != nil {
		return fmt.Errorf("failed to open temp file: %w", err)
	}

	if _, err := io.Copy(wf, resp.Body); err != nil {
		return fmt.Errorf("failed to write to temp file: %w", err)
	}

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
