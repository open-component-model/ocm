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
	"path/filepath"
	"strings"
	"time"

	"github.com/fluxcd/pkg/ssa"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/cli-runtime/pkg/genericclioptions"

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
	Namespace      string
	ControllerName string
	Timeout        time.Duration
	Version        string
	BaseURL        string
	ReleaseAPIURL  string
	DryRun         bool
}

var _ utils.OCMCommand = (*Command)(nil)

// NewCommand creates a new controller cdommand.
func NewCommand(ctx clictx.Context, names ...string) *cobra.Command {
	return utils.SetupCommand(&Command{BaseCommand: utils.NewBaseCommand(ctx)}, utils.Names(Names, names...)...)
}

func (o *Command) ForName(name string) *cobra.Command {
	return &cobra.Command{
		Use:   "install controller {--version v0.0.1}",
		Short: "Install either a specific or latest version of the ocm-controller.",
	}
}

func (o *Command) AddFlags(set *pflag.FlagSet) {
	set.StringVarP(&o.Version, "version", "v", "latest", "the version of the controller to install")
	set.StringVarP(&o.BaseURL, "base-url", "u", "https://github.com/open-component-model/ocm-controller/releases", "the base url to the ocm-controller's release page")
	set.StringVarP(&o.ReleaseAPIURL, "release-api-url", "a", "https://api.github.com/repos/open-component-model/ocm-controller/releases", "the base url to the ocm-controller's API release page")
	set.StringVarP(&o.ControllerName, "controller-name", "c", "ocm-controller", "name of the controller that's used for status check")
	set.StringVarP(&o.Namespace, "namespace", "n", "ocm-system", "the namespace into which the controller is installed")
	set.DurationVarP(&o.Timeout, "timeout", "t", 1*time.Minute, "maximum time to wait for deployment to be ready")
	set.BoolVarP(&o.DryRun, "dry-run", "d", false, "if enabled, prints the downloaded manifest file")
}

func (o *Command) Complete(args []string) error {
	return nil
}

func (o *Command) Run() error {
	out.Outf(o.Context, "► installing ocm-controller with version %s\n", o.Version)
	version := o.Version
	if version == "latest" {
		latest, err := o.GetLatestVersion()
		if err != nil {
			return fmt.Errorf("✗ failed to retrieve latest version for ocm-controller: %s", err)
		}
		out.Outf(o.Context, "► got latest version %q\n", latest)
		version = latest
	} else {
		exists, err := o.ExistingVersion(version)
		if err != nil {
			return fmt.Errorf("✗ failed to check if version exists: %w", err)
		}
		if !exists {
			return fmt.Errorf("✗ version %q does not exist", version)
		}
	}

	temp, err := os.MkdirTemp("", "ocm-controller-download")
	if err != nil {
		return fmt.Errorf("✗ failed to create temp folder: %w", err)
	}
	defer os.RemoveAll(temp)

	if err := o.fetch(context.Background(), version, temp); err != nil {
		return fmt.Errorf("✗ failed to download install.yaml file: %w", err)
	}

	path := filepath.Join(temp, "install.yaml")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("✗ failed to find install.yaml file at location: %w", err)
	}
	out.Outf(o.Context, "✔ successfully fetched install file\n")
	if o.DryRun {
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("✗ failed to read install.yaml file at location: %w", err)
		}
		out.Outf(o.Context, string(content))
		return nil
	}
	out.Outf(o.Context, "► applying to cluster...\n")

	kubeconfigArgs := genericclioptions.NewConfigFlags(false)
	sm, err := NewResourceManager(kubeconfigArgs)
	if err != nil {
		return fmt.Errorf("✗ failed to create resource manager: %w", err)
	}

	objects, err := readObjects(path)
	if err != nil {
		return fmt.Errorf("✗ failed to construct objects to apply: %w", err)
	}

	if _, err := sm.ApplyAllStaged(context.Background(), objects, ssa.DefaultApplyOptions()); err != nil {
		return fmt.Errorf("✗ failed to apply manifests: %w", err)
	}

	out.Outf(o.Context, "► waiting for ocm deployment to be ready\n")
	if err = sm.Wait(objects, ssa.DefaultWaitOptions()); err != nil {
		return fmt.Errorf("✗ failed to wait for objects to be ready: %w", err)
	}

	out.Outf(o.Context, "✔ ocm-controller successfully installed\n")
	return nil
}

// GetLatestVersion calls the GitHub API and returns the latest released version.
func (o *Command) GetLatestVersion() (string, error) {
	c := http.DefaultClient
	c.Timeout = 15 * time.Second

	res, err := c.Get(o.ReleaseAPIURL + "/latest")
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

// ExistingVersion calls the GitHub API to confirm the given version does exist.
func (o *Command) ExistingVersion(version string) (bool, error) {
	if !strings.HasPrefix(version, "v") {
		version = "v" + version
	}

	ghURL := fmt.Sprintf(o.ReleaseAPIURL+"/tags/%s", version)
	c := http.DefaultClient
	c.Timeout = 15 * time.Second

	res, err := c.Get(ghURL)
	if err != nil {
		return false, fmt.Errorf("GitHub API call failed: %w", err)
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

func (o *Command) fetch(ctx context.Context, version, dir string) error {
	ghURL := fmt.Sprintf("%s/latest/download/install.yaml", o.BaseURL)
	if strings.HasPrefix(version, "v") {
		ghURL = fmt.Sprintf("%s/download/%s/install.yaml", o.BaseURL, version)
	}

	req, err := http.NewRequest("GET", ghURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create HTTP request for %s, error: %w", ghURL, err)
	}

	// download
	resp, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("failed to download manifests.tar.gz from %s, error: %w", ghURL, err)
	}
	defer resp.Body.Close()

	// check response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download manifests.tar.gz from %s, status: %s", ghURL, resp.Status)
	}

	wf, err := os.OpenFile(filepath.Join(dir, "install.yaml"), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		return fmt.Errorf("failed to open temp file: %w", err)
	}

	if _, err := io.Copy(wf, resp.Body); err != nil {
		return fmt.Errorf("failed to write to temp file: %w", err)
	}

	return nil
}
