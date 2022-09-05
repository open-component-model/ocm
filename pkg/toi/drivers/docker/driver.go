// Copyright 2022 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package docker

import (
	"archive/tar"
	"context"
	"fmt"
	"io"
	"os"
	unix_path "path"
	"strconv"
	"strings"

	"github.com/docker/cli/cli/command"
	"github.com/docker/distribution/reference"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/docker/registry"
	"github.com/mitchellh/copystructure"
	"github.com/pkg/errors"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/toi/install"
)

const (
	OptionQuiet            = "DOCKER_DRIVER_QUIET"
	OptionCleanup          = "CLEANUP_CONTAINERS"
	OptionPullPolicy       = "PULL_POLICY"
	PullPolicyAlways       = "Always"
	PullPolicyNever        = "Never"
	PullPolicyIfNotPresent = "IfNotPresent"

	trueAsString  = "true"
	falseAsString = "false"
)

// Driver is capable of running Docker invocation images using Docker itself.
type Driver struct {
	config map[string]string
	// If true, this will not actually run Docker
	Simulate                   bool
	dockerCli                  command.Cli
	dockerConfigurationOptions []ConfigurationOption
	containerOut               io.Writer
	containerErr               io.Writer
	containerHostCfg           container.HostConfig
	containerCfg               container.Config
}

var _ install.Driver = (*Driver)(nil)

func New() install.Driver {
	return &Driver{}
}

// GetContainerConfig returns a copy of the container configuration
// used by the driver during container exec.
func (d *Driver) GetContainerConfig() (container.Config, error) {
	cpy, err := copystructure.Copy(d.containerCfg)
	if err != nil {
		return container.Config{}, err
	}

	cfg, ok := cpy.(container.Config)
	if !ok {
		return container.Config{}, errors.New("unable to process container config")
	}

	return cfg, nil
}

// GetContainerHostConfig returns a copy of the container host configuration
// used by the driver during container exec.
func (d *Driver) GetContainerHostConfig() (container.HostConfig, error) {
	cpy, err := copystructure.Copy(d.containerHostCfg)
	if err != nil {
		return container.HostConfig{}, err
	}

	cfg, ok := cpy.(container.HostConfig)
	if !ok {
		return container.HostConfig{}, errors.New("unable to process container host config")
	}

	return cfg, nil
}

// SetConfig sets Docker driver configuration.
func (d *Driver) SetConfig(settings map[string]string) error {
	// Set default and provide feedback on acceptable input values.
	value, ok := settings[OptionCleanup]
	if !ok {
		settings[OptionCleanup] = trueAsString
	} else if value != trueAsString && value != falseAsString {
		return fmt.Errorf("config variable %s has unexpected value %q. Supported values are 'true', 'false', or unset", OptionCleanup, value)
	}

	value, ok = settings[OptionPullPolicy]
	if ok {
		if value != PullPolicyAlways && value != PullPolicyIfNotPresent && value != PullPolicyNever {
			return fmt.Errorf("config variable %s has unexpected value %q. Supported values are '%s', '%s', '%s' , or unset", OptionPullPolicy, value, PullPolicyAlways, PullPolicyIfNotPresent, PullPolicyNever)
		}
	}
	d.config = settings
	return nil
}

// SetDockerCli makes the driver use an already initialized cli.
func (d *Driver) SetDockerCli(dockerCli command.Cli) {
	d.dockerCli = dockerCli
}

// SetContainerOut sets the container output stream.
func (d *Driver) SetContainerOut(w io.Writer) {
	d.containerOut = w
}

// SetContainerErr sets the container error stream.
func (d *Driver) SetContainerErr(w io.Writer) {
	d.containerErr = w
}

func pullImage(ctx context.Context, cli command.Cli, image string) error {
	ref, err := reference.ParseNormalizedNamed(image)
	if err != nil {
		return fmt.Errorf("unable to parse normalized name: %w", err)
	}

	// Resolve the Repository name from fqn to RepositoryInfo
	repoInfo, err := registry.ParseRepositoryInfo(ref)
	if err != nil {
		return fmt.Errorf("unable to parse repository info: %w", err)
	}

	authConfig := command.ResolveAuthConfig(ctx, cli, repoInfo.Index)

	encodedAuth, err := command.EncodeAuthToBase64(authConfig)
	if err != nil {
		return fmt.Errorf("unable encode auth: %w", err)
	}

	options := types.ImagePullOptions{
		RegistryAuth: encodedAuth,
	}

	responseBody, err := cli.Client().ImagePull(ctx, image, options)
	if err != nil {
		return fmt.Errorf("unable to pull image: %w", err)
	}

	defer responseBody.Close()

	// passing isTerm = false here because of https://github.com/Nvveen/Gotty/pull/1
	err = jsonmessage.DisplayJSONMessagesStream(responseBody, cli.Out(), cli.Out().FD(), false, nil)
	if err != nil {
		return fmt.Errorf("unable to display json message: %w", err)
	}

	return nil
}

func (d *Driver) initializeDockerCli() (command.Cli, error) {
	if d.dockerCli != nil {
		return d.dockerCli, nil
	}

	cli, err := GetDockerClient()
	if err != nil {
		return nil, err
	}

	if d.config[OptionQuiet] == "1" {
		cli.Apply(command.WithCombinedStreams(io.Discard))
	}

	d.dockerCli = cli
	return cli, nil
}

func (d *Driver) Exec(op *install.Operation) (*install.OperationResult, error) {
	ctx := context.Background()

	cli, err := d.initializeDockerCli()
	if err != nil {
		return nil, err
	}

	if d.Simulate {
		return nil, nil
	}
	if d.config[OptionPullPolicy] == PullPolicyAlways {
		if err := pullImage(ctx, cli, op.Image.Ref); err != nil {
			return nil, err
		}
	}

	ii, err := d.inspectImage(ctx, op.Image.Ref)
	if err != nil {
		return nil, err
	}

	err = d.validateImageDigest(op.Image, ii.RepoDigests)
	if err != nil {
		return nil, errors.Wrap(err, "image digest validation failed")
	}

	if err := d.setConfigurationOptions(op); err != nil {
		return nil, err
	}

	resp, err := cli.Client().ContainerCreate(ctx, &d.containerCfg, &d.containerHostCfg, nil, nil, "")
	if err != nil {
		return nil, fmt.Errorf("cannot create container: %w", err)
	}

	if d.config[OptionCleanup] == trueAsString {
		defer cli.Client().ContainerRemove(ctx, resp.ID, types.ContainerRemoveOptions{})
	}

	containerUID := getContainerUserID(ii.Config.User)
	tarContent, err := generateTar(op.Files, containerUID)
	if err != nil {
		return nil, fmt.Errorf("error staging files: %w", err)
	}
	options := types.CopyToContainerOptions{
		AllowOverwriteDirWithFile: false,
	}
	// This copies the tar to the root of the container. The tar has been assembled using the
	// path from the given file, starting at the /.
	err = cli.Client().CopyToContainer(ctx, resp.ID, "/", tarContent, options)
	if err != nil {
		return nil, fmt.Errorf("error copying to / in container: %w", err)
	}

	attach, err := cli.Client().ContainerAttach(ctx, resp.ID, types.ContainerAttachOptions{
		Stream: true,
		Stdout: true,
		Stderr: true,
		Logs:   true,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve logs: %w", err)
	}
	var (
		stdout io.Writer = os.Stdout
		stderr io.Writer = os.Stderr
	)
	if d.containerOut != nil {
		stdout = d.containerOut
	} else if op.Out != nil {
		stdout = op.Out
	}
	if d.containerErr != nil {
		stderr = d.containerErr
	} else if op.Err != nil {
		stderr = op.Err
	}
	go func() {
		defer attach.Close()
		for {
			_, err = stdcopy.StdCopy(stdout, stderr, attach.Reader)
			if err != nil {
				break
			}
		}
	}()

	if err = cli.Client().ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return nil, fmt.Errorf("cannot start container: %w", err)
	}
	statusc, errc := cli.Client().ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errc:
		if err != nil {
			opResult, fetchErr := d.fetchOutputs(ctx, resp.ID, op)
			return opResult, containerError("error in container", err, fetchErr)
		}
	case s := <-statusc:
		if s.StatusCode == 0 {
			return d.fetchOutputs(ctx, resp.ID, op)
		}
		if s.Error != nil {
			opResult, fetchErr := d.fetchOutputs(ctx, resp.ID, op)
			return opResult, containerError(fmt.Sprintf("container exit code: %d, message", s.StatusCode), err, fetchErr)
		}
		opResult, fetchErr := d.fetchOutputs(ctx, resp.ID, op)
		return opResult, containerError(fmt.Sprintf("container exit code: %d, message", s.StatusCode), err, fetchErr)
	}
	opResult, fetchErr := d.fetchOutputs(ctx, resp.ID, op)
	if fetchErr != nil {
		return opResult, fmt.Errorf("fetching outputs failed: %w", fetchErr)
	}
	return opResult, err
}

// getContainerUserID determines the user id that the container will execute as
// based on the image's configured user. Defaults to 0 (root) if a user id is not set.
func getContainerUserID(user string) int {
	if user != "" {
		// Only look at the user, strip off a group if one was specified with USER uid:gid
		if uid, err := strconv.Atoi(strings.Split(user, ":")[0]); err == nil {
			return uid
		}
	}
	return 0
}

// ApplyConfigurationOptions applies the configuration options set on the driver by the user.
func (d *Driver) ApplyConfigurationOptions() error {
	for _, opt := range d.dockerConfigurationOptions {
		if err := opt(&d.containerCfg, &d.containerHostCfg); err != nil {
			return fmt.Errorf("unable to apply docker configuration: %w", err)
		}
	}

	return nil
}

// setConfigurationOptions initializes the container and host configuration options on the driver,
// combining the default configuration with any overrides set by the user.
func (d *Driver) setConfigurationOptions(op *install.Operation) error {
	var env []string
	for k, v := range op.Environment {
		env = append(env, fmt.Sprintf("%s=%v", k, v))
	}

	d.containerCfg = container.Config{
		Image:        op.Image.Ref,
		Env:          env,
		Cmd:          []string{op.Action, op.ComponentVersion},
		AttachStderr: true,
		AttachStdout: true,
	}

	d.containerHostCfg = container.HostConfig{}

	if err := d.ApplyConfigurationOptions(); err != nil {
		return fmt.Errorf("failed to apply configuration: %w", err)
	}

	return nil
}

func containerError(containerMessage string, containerErr, fetchErr error) error {
	if fetchErr != nil {
		return fmt.Errorf("%s: %v. fetching outputs failed: %w", containerMessage, containerErr, fetchErr)
	}
	return fmt.Errorf("%s: %w", containerMessage, containerErr)
}

// fetchOutputs takes a context and a container ID; it copies the PathOutputs directory from that container.
// The goal is to collect all the files in the directory (recursively) and put them in a flat map of path to contents.
// This map will be inside the OperationResult. When fetchOutputs returns an error, it may also return partial results.
func (d *Driver) fetchOutputs(ctx context.Context, container string, op *install.Operation) (*install.OperationResult, error) {
	opResult := &install.OperationResult{
		Outputs: map[string][]byte{},
	}
	// The PathOutputs directory probably only exists if outputs are created. In the
	// case there are no outputs defined on the operation, there probably are none to copy
	// and we should return early.
	if len(op.Outputs) == 0 {
		return opResult, nil
	}
	ioReader, _, err := d.dockerCli.Client().CopyFromContainer(ctx, container, install.PathOutputs)
	if err != nil {
		return nil, fmt.Errorf("error copying outputs from container: %w", err)
	}
	tarReader := tar.NewReader(ioReader)
	header, err := tarReader.Next()
	// io.EOF pops us out of loop on successful run.
	for err == nil {
		// skip directories because we're gathering file contents
		if header.FileInfo().IsDir() {
			header, err = tarReader.Next()
			continue
		}

		var contents []byte
		// CopyFromContainer strips prefix above outputs directory.
		name := strings.TrimPrefix(header.Name, "outputs/")
		outputName, shouldCapture := op.Outputs[name]
		if shouldCapture {
			contents, err = io.ReadAll(tarReader)
			if err != nil {
				return opResult, fmt.Errorf("error while reading %q from outputs tar: %w", header.Name, err)
			}
			opResult.Outputs[outputName] = contents
		}

		header, err = tarReader.Next()
	}

	if !errors.Is(err, io.EOF) {
		return opResult, err
	}

	return opResult, nil
}

// generateTar creates a tarfile containing the specified files, with the owner
// set to the uid that the container runs as so that it is guaranteed to have
// read access to the files we copy into the container.
func generateTar(files map[string]accessio.BlobAccess, uid int) (io.Reader, error) {
	r, w := io.Pipe()
	tw := tar.NewWriter(w)
	for path := range files {
		if unix_path.IsAbs(path) {
			return nil, fmt.Errorf("destination path %s should be a relative unix path", path)
		}
	}
	go func() {
		have := map[string]bool{}

		for path, content := range files {
			path = unix_path.Join(install.PathInputs, path)
			// Write a header for the parent directories so that newly created intermediate directories are accessible by the user
			dir := path
			for dir != "/" {
				dir = unix_path.Dir(dir)
				if !have[dir] {
					dirHdr := &tar.Header{
						Typeflag: tar.TypeDir,
						Name:     dir,
						Mode:     0o700,
						Uid:      uid,
						Size:     0,
					}
					tw.WriteHeader(dirHdr)
					have[dir] = true
				}
			}

			// Grant access to just the owner (container user), so that files can be read by the container
			fildHdr := &tar.Header{
				Typeflag: tar.TypeReg,
				Name:     path,
				Mode:     0o600,
				Size:     content.Size(),
				Uid:      uid,
			}
			tw.WriteHeader(fildHdr)
			reader, _ := content.Reader()
			io.Copy(tw, reader)
		}
		w.Close()
	}()
	return r, nil
}

// ConfigurationOption is an option used to customize docker driver container and host config.
type ConfigurationOption func(*container.Config, *container.HostConfig) error

// inspectImage inspects the operation image and returns an object of types.ImageInspect,
// pulling the image if not found locally.
func (d *Driver) inspectImage(ctx context.Context, image string) (types.ImageInspect, error) {
	ii, _, err := d.dockerCli.Client().ImageInspectWithRaw(ctx, image)
	switch {
	case client.IsErrNotFound(err):
		fmt.Fprintf(d.dockerCli.Err(), "Unable to find image '%s' locally\n", image)
		if d.config[OptionPullPolicy] == PullPolicyNever {
			return ii, errors.Wrapf(err, "image %s not found", image)
		}
		if err := pullImage(ctx, d.dockerCli, image); err != nil {
			return ii, err
		}
		if ii, _, err = d.dockerCli.Client().ImageInspectWithRaw(ctx, image); err != nil {
			return ii, errors.Wrapf(err, "cannot inspect image %s", image)
		}
	case err != nil:
		return ii, errors.Wrapf(err, "cannot inspect image %s", image)
	}

	return ii, nil
}

// validateImageDigest validates the operation image digest, if exists, against
// the supplied repoDigests.
func (d *Driver) validateImageDigest(image install.Image, repoDigests []string) error {
	if image.Digest == "" {
		return nil
	}

	if len(repoDigests) == 0 {
		return fmt.Errorf("image %s has no repo digests", image)
	}

	for _, repoDigest := range repoDigests {
		// RepoDigests are of the form 'imageName@sha256:<sha256>' or imageName:<tag>
		// We only care about the ones in digest form
		ref, err := reference.ParseNormalizedNamed(repoDigest)
		if err != nil {
			return fmt.Errorf("unable to parse repo digest %s", repoDigest)
		}

		digestRef, ok := ref.(reference.Digested)
		if !ok {
			continue
		}

		digest := digestRef.Digest().String()

		// image.Digest is the digest of the original invocation image defined in the bundle.
		// It persists even when the bundle's invocation image has been relocated.
		if digest == image.Digest {
			return nil
		}
	}

	return fmt.Errorf("content digest mismatch: image %s was defined with the digest %s, but no matching repoDigest was found upon inspecting the image", image.Ref, image.Digest)
}
