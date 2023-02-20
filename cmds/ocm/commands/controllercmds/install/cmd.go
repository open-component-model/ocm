// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package install

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/open-component-model/ocm/cmds/ocm/commands/controllercmds/names"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
	"github.com/open-component-model/ocm/pkg/out"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	Names = names.Controller
	Verb  = verbs.Install
)

type Command struct {
	utils.BaseCommand
	Version       string
	BaseURL       string
	ReleaseAPIURL string
	DryRun        bool
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

	kubectlArgs := []string{"apply", "-f", path}
	if _, err := ExecKubectlCommand(context.Background(), ModeOS, "", "", kubectlArgs...); err != nil {
		return fmt.Errorf("✗ failed to apply manifest to cluster: %w", err)
	}

	out.Outf(o.Context, "✔ successfully applied manifests to cluster\n")
	out.Outf(o.Context, "◎ waiting for pod to become Ready\n")
	kubectlArgs = []string{"wait", "-l", "app=ocm-controller", "-n", "ocm-system", "--for", "condition=Ready", "--timeout=90s", "pod"}
	if _, err := ExecKubectlCommand(context.Background(), ModeOS, "", "", kubectlArgs...); err != nil {
		return fmt.Errorf("✗ failed to wait for pod to be ready: %w", err)
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

type ExecMode string

const (
	ModeOS       ExecMode = "os.stderr|stdout"
	ModeStderrOS ExecMode = "os.stderr"
	ModeCapture  ExecMode = "capture.stderr|stdout"
)

func ExecKubectlCommand(ctx context.Context, mode ExecMode, kubeConfigPath string, kubeContext string, args ...string) (string, error) {
	var stdoutBuf, stderrBuf bytes.Buffer

	if kubeConfigPath != "" && len(filepath.SplitList(kubeConfigPath)) == 1 {
		args = append(args, "--kubeconfig="+kubeConfigPath)
	}

	if kubeContext != "" {
		args = append(args, "--context="+kubeContext)
	}

	c := exec.CommandContext(ctx, "kubectl", args...)

	if mode == ModeStderrOS {
		c.Stderr = io.MultiWriter(os.Stderr, &stderrBuf)
	}
	if mode == ModeOS {
		c.Stdout = io.MultiWriter(os.Stdout, &stdoutBuf)
		c.Stderr = io.MultiWriter(os.Stderr, &stderrBuf)
	}

	if mode == ModeStderrOS || mode == ModeOS {
		if err := c.Run(); err != nil {
			return "", err
		} else {
			return "", nil
		}
	}

	if mode == ModeCapture {
		c.Stdout = &stdoutBuf
		c.Stderr = &stderrBuf
		if err := c.Run(); err != nil {
			return stderrBuf.String(), err
		} else {
			return stdoutBuf.String(), nil
		}
	}

	return "", nil
}
