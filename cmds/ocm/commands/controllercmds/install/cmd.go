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
	common2 "github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	Names   = names.Controller
	Verb    = verbs.Install
	BaseURL = "https://github.com/open-component-model/ocm-controller/releases"
)

type Command struct {
	utils.BaseCommand
	Version string
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
}

func (o *Command) Complete(args []string) error {
	return nil
}

func (o *Command) Run() error {
	p := common2.NewPrinter(os.Stdout)
	p.Printf("► installing ocm-controller with version %s\n", o.Version)
	version := o.Version
	if version == "latest" {
		latest, err := GetLatestVersion()
		if err != nil {
			p.Printf("✗ failed to retrieve latest version for ocm-controller: %s", err.Error())
			return err
		}
		p.Printf("► got latest version %q\n", latest)
		version = latest
	} else {
		exists, err := ExistingVersion(version)
		if err != nil {
			p.Printf("✗ failed to check if version exists: %s", err.Error())
			return err
		}
		if !exists {
			p.Printf("✗ version %q does not exist\n", version)
			return err
		}
	}

	temp, err := os.MkdirTemp("", "ocm-controller-download")
	if err != nil {
		p.Printf("✗ failed to create temp folder: %w", err)
		return err
	}
	defer os.RemoveAll(temp)

	if err := fetch(context.Background(), version, temp); err != nil {
		p.Printf("✗ failed to download install.yaml file: %w", err)
		return err
	}

	path := filepath.Join(temp, "install.yaml")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		p.Printf("✗ failed to find install.yaml file at location: %s", path)
		return err
	}
	p.Printf("✔ successfully fetched install file to %s\n", path)
	p.Printf("► applying to cluster...\n")

	kubectlArgs := []string{"apply", "-f", path}
	if _, err := ExecKubectlCommand(context.Background(), ModeOS, "", "", kubectlArgs...); err != nil {
		p.Printf("✗ failed to apply manifest to cluster: %w", err)
		return err
	}

	p.Printf("✔ successfully applied manifests to cluster\n")
	p.Printf("◎ waiting for pod to become Ready\n")
	kubectlArgs = []string{"wait", "-l", "app=ocm-controller", "-n", "ocm-system", "--for", "condition=Ready", "--timeout=90s", "pod"}
	if _, err := ExecKubectlCommand(context.Background(), ModeOS, "", "", kubectlArgs...); err != nil {
		p.Printf("✗ failed to wait for pod to be ready: %w", err)
		return err
	}

	p.Printf("✔ ocm-controller successfully installed\n")
	return nil
}

// GetLatestVersion calls the GitHub API and returns the latest released version.
func GetLatestVersion() (string, error) {
	ghURL := "https://api.github.com/repos/open-component-model/ocm-controller/releases/latest"
	c := http.DefaultClient
	c.Timeout = 15 * time.Second

	res, err := c.Get(ghURL)
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
func ExistingVersion(version string) (bool, error) {
	if !strings.HasPrefix(version, "v") {
		version = "v" + version
	}

	ghURL := fmt.Sprintf("https://api.github.com/repos/fluxcd/flux2/releases/tags/%s", version)
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

func fetch(ctx context.Context, version, dir string) error {
	ghURL := fmt.Sprintf("%s/latest/download/install.yaml", BaseURL)
	if strings.HasPrefix(version, "v") {
		ghURL = fmt.Sprintf("%s/download/%s/install.yaml", BaseURL, version)
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
