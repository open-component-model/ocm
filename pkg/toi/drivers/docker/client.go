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
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/docker/cli/cli/command"
	cliconfig "github.com/docker/cli/cli/config"
	cliflags "github.com/docker/cli/cli/flags"
	"github.com/docker/go-connections/tlsconfig"
)

const (
	// DockerTLSVerifyEnvVar is the Docker environment variable that indicates that
	// Docker socket is protected with TLS.
	DockerTLSVerifyEnvVar = "DOCKER_TLS_VERIFY"

	// DockerCertPathEnvVar is the Docker environment variable that specifies a
	// custom path to the TLS certificates for the Docker socket.
	DockerCertPathEnvVar = "DOCKER_CERT_PATH"
)

// GetDockerClient creates a Docker CLI client that uses the user's Docker configuration
// such as environment variables and the Docker home directory to initialize the client.
func GetDockerClient() (*command.DockerCli, error) {
	cli, err := command.NewDockerCli()
	if err != nil {
		return nil, fmt.Errorf("could not create new docker client: %w", err)
	}
	opts := buildDockerClientOptions()
	if err = cli.Initialize(opts); err != nil {
		return nil, fmt.Errorf("error initializing docker client: %w", err)
	}
	return cli, nil
}

// manually handle DOCKER_TLS_VERIFY and DOCKER_CERT_PATH because the docker cli
// library only binds these values when initializing its cli flags. There isn't
// other parts of the library that we can take advantage of to get these values
// for "free".
//
// DOCKER_HOST however is retrieved dynamically later so that doesn't
// require additional configuration.
func buildDockerClientOptions() *cliflags.ClientOptions {
	cliOpts := cliflags.NewClientOptions()
	cliOpts.ConfigDir = cliconfig.Dir()

	// Check if TLS is enabled Docker configures TLS settings if DOCKER_TLS_VERIFY is
	// set to anything, so it could be false and that still means we should use TLS
	// (but don't check the certs).
	tlsVerify, tlsConfigured := os.LookupEnv(DockerTLSVerifyEnvVar)
	if tlsConfigured && tlsVerify != "" {
		cliOpts.Common.TLS = true

		// Check if we should verify certs or allow self-signed certs (insecure)
		verify, _ := strconv.ParseBool(tlsVerify)
		cliOpts.Common.TLSVerify = verify

		// Check if the TLS certs have been overridden
		var certPath string
		if certPathOverride, ok := os.LookupEnv(DockerCertPathEnvVar); ok && certPathOverride != "" {
			certPath = certPathOverride
		} else {
			certPath = cliOpts.ConfigDir
		}

		cliOpts.Common.TLSOptions = &tlsconfig.Options{
			CAFile:   filepath.Join(certPath, cliflags.DefaultCaFile),
			CertFile: filepath.Join(certPath, cliflags.DefaultCertFile),
			KeyFile:  filepath.Join(certPath, cliflags.DefaultKeyFile),
		}
	}

	return cliOpts
}
