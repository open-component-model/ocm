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

package main

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/containerd/containerd/errdefs"
	"github.com/containerd/containerd/log"
	"github.com/containerd/containerd/reference"
	"github.com/containerd/containerd/remotes"
	"github.com/containerd/containerd/remotes/docker"
	"github.com/containerd/containerd/remotes/docker/config"
	config2 "github.com/docker/cli/cli/config"
	"github.com/docker/cli/cli/config/credentials"
	"github.com/gardener/ocm/pkg/common/accessio"
	"github.com/gardener/ocm/pkg/oci/artdesc"
	"github.com/sirupsen/logrus"
)

const MIME_OCTET = "application/octet-stream"

func fetch(ctx context.Context, msg string, f remotes.Fetcher, desc artdesc.Descriptor) {
	fmt.Printf("*** fetch %s %s\n", desc.MediaType, desc.Digest)
	read, err := f.Fetch(ctx, desc)
	if err != nil {
		fmt.Printf("fetch %s failed: %s\n", msg, err)
		os.Exit(1)
	}

	data, err := ioutil.ReadAll(read)
	read.Close()
	if err != nil {
		fmt.Printf("read %s failed: %s\n", msg, err)
		os.Exit(1)
	}
	fmt.Printf("*** %s: %s\n", msg, string(data))
}

func push(ctx context.Context, msg string, p remotes.Pusher, blob accessio.BlobAccess) {
	desc := *artdesc.DefaultBlobDescriptor(blob)
	key := remotes.MakeRefKey(ctx, desc)
	fmt.Printf("*** push %s %s: %s\n", desc.MediaType, desc.Digest, key)
	write, err := p.Push(ctx, desc)
	if err != nil {
		if errdefs.IsAlreadyExists(err) {
			fmt.Printf("%s already exists\n", msg)
			return
		}
		fmt.Printf("push %s failed: %s\n", msg, err)
		os.Exit(1)
	}
	read, err := blob.Reader()
	defer read.Close()
	_, err = io.Copy(write, read)
	if err != nil {
		fmt.Printf("copy %s failed: %s\n", msg, err)
		os.Exit(1)
	}
	err = write.Commit(ctx, desc.Size, desc.Digest)
	if err != nil {
		fmt.Printf("commit %s failed: %s\n", msg, err)
		os.Exit(1)
	}
}

func main() {
	ctx := context.Background()
	logger := logrus.New()
	logger.Level = logrus.ErrorLevel
	ctx = log.WithLogger(ctx, logrus.NewEntry(logger))

	cfg := config2.LoadDefaultConfigFile(os.Stderr)
	if cfg == nil {
		fmt.Printf("failed to load dockercfg\n")
		os.Exit(1)
	}
	for n, c := range cfg.GetAuthConfigs() {
		fmt.Printf("%s (%s): %s\n", credentials.ConvertToHostname(n), n, c)
	}
	opts := docker.ResolverOptions{
		Hosts: config.ConfigureHosts(context.Background(), config.HostOptions{
			Credentials: func(host string) (string, string, error) {
				a, err := cfg.GetAuthConfig(host)
				if err == nil {
					fmt.Printf("ath: %s\n", a)
					p := a.Password
					if a.RegistryToken != "" {
						p = a.RegistryToken
					}
					return a.Username, p, err
				}
				return "", "", err
			},
		}),
	}

	r := docker.NewResolver(opts)

	ref := "ghcr.io/mandelsoft/cnudie/component-descriptors/github.com/mandelsoft/pause:0.1-dev"
	ref = "eu.gcr.io/sap-se-gcr-k8s-public/eu_gcr_io/gardener-project/cert-controller-manager@sha256:dd28a472a488aaef4c5bfe8de02dc225e97d85b71160f30edd1f2b0c83ffcf8a"
	refspec, err := reference.Parse(ref)
	if err != nil {
		fmt.Printf("ilvalid ref: %s\n", err)
		os.Exit(1)
	}
	fmt.Printf("%s: locator %s, object %s\n", refspec, refspec.Locator, refspec.Object)

	name, desc, err := r.Resolve(ctx, ref)
	//name, desc, err := r.Resolve(context.Background(), "gcr.io/mandelsoft/cnudie/component-descriptors/github.com/mandelsoft/pause:0.1-dev")
	//name, desc, err := r.Resolve(context.Background(), "mandelsoft/kubelink:latest")

	if err != nil {
		fmt.Printf("failed: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("%s: digest: %s\n", name, desc.Digest)

	f, err := r.Fetcher(ctx, name)

	if err != nil {
		fmt.Printf("failed: %s\n", err)
		os.Exit(1)
	}

	blob := accessio.BlobAccessForString(MIME_OCTET, "testdata4")
	fetch(ctx, "manifest", f, desc)
	//	desc2:= desc
	//	desc2.MediaType=artdesc.MediaTypeImageIndex+", "+artdesc.MediaTypeImageManifest
	//	fetch(ctx, "fake manifest", f, desc2)

	p, err := r.Pusher(ctx, ref)
	if err != nil {
		fmt.Printf("get pusher failed: %s\n", err)
		os.Exit(1)
	}

	push(ctx, "blob", p, blob)

	desc = *artdesc.DefaultBlobDescriptor(blob)
	desc.MediaType = ""
	fetch(ctx, "blob", f, desc)

}
