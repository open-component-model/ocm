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
	"fmt"
	"os"

	"github.com/gardener/ocm/pkg/common/accessio"
	"github.com/gardener/ocm/pkg/credentials"
	"github.com/gardener/ocm/pkg/credentials/repositories/dockerconfig"
	"github.com/gardener/ocm/pkg/oci"
	"github.com/gardener/ocm/pkg/oci/ociutils"
	"github.com/gardener/ocm/pkg/oci/repositories/ocireg"
	_ "github.com/gardener/ocm/pkg/ocm"
)

const MIME_OCTET = "application/octet-stream"

func Error(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "ERROR: "+msg, args...)
	os.Exit(1)
}

func handleError(err error, msg string, args ...interface{}) {
	if err != nil {
		Error("%s; %s", fmt.Sprintf(msg, args...), err)
	}
}

func setupCredentials() {
	ctx := credentials.DefaultContext()

	spec := dockerconfig.NewRepositorySpec(os.Getenv("HOME") + "/.docker/config.json").WithConsumerPropagation(true)
	repo, err := ctx.RepositoryForSpec(spec)
	handleError(err, "setup credentials")
	_ = repo
}
func main() {

	setupCredentials()

	ctx := oci.DefaultContext()

	spec := ocireg.NewRepositorySpec("ghcr.io")

	repo, err := ctx.RepositoryForSpec(spec)
	handleError(err, "get repo")

	ns, err := repo.LookupNamespace("mandelsoft/cnudie/component-descriptors/github.com/mandelsoft/pause")
	handleError(err, "lookup namepsace")

	art, err := ns.GetArtefact("0.1-dev")
	handleError(err, "lookup artefact")

	fmt.Println(ociutils.PrintArtefact(art))
	_ = art

	blob := accessio.BlobAccessForString(MIME_OCTET, "testdata")
	err = ns.AddBlob(blob)
	handleError(err, "add blob")

	bd, err := ns.GetBlobData(blob.Digest())
	data, err := bd.Get()
	handleError(err, "read blob")
	fmt.Println(string(data))

	b, _ := art.Blob()
	err = ns.AddTags(b.Digest(), "0.1.beta")
	handleError(err, "add tag")
}
