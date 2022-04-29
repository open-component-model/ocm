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
	"github.com/open-component-model/ocm/cmds/ocm/clictx"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/inputs"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/contexts/oci/repositories/artefactset"
	"github.com/open-component-model/ocm/pkg/contexts/oci/repositories/docker"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

type Handler struct{}

var _ inputs.InputHandler = (*Handler)(nil)

func (h *Handler) Validate(fldPath *field.Path, ctx clictx.Context, input *inputs.BlobInput, inputFilePath string) field.ErrorList {
	allErrs := inputs.ForbidFileInfo(fldPath, input)
	pathField := fldPath.Child("path")
	if input.Path == "" {
		allErrs = append(allErrs, field.Required(pathField, "path is required for input"))
	} else {
		_, _, err := docker.ParseGenericRef(input.Path)
		if err != nil {
			allErrs = append(allErrs, field.Invalid(pathField, input.Path, err.Error()))

		}
	}
	return allErrs
}

func (h *Handler) GetBlob(ctx clictx.Context, input *inputs.BlobInput, inputFilePath string) (accessio.TemporaryBlobAccess, string, error) {
	locator, version, err := docker.ParseGenericRef(input.Path)
	if err != nil {
		return nil, "", err
	}
	spec := docker.NewRepositorySpec()
	repo, err := ctx.OCIContext().RepositoryForSpec(spec)
	if err != nil {
		return nil, "", err
	}
	ns, err := repo.LookupNamespace(locator)
	if err != nil {
		return nil, "", err
	}

	blob, err := artefactset.SynthesizeArtefactBlob(ns, version)
	if err != nil {
		return nil, "", err
	}
	return blob, locator, nil
}


func (h *Handler) Usage() string {
	return `
- <code>docker</code>

  The path must denote an image tag that can be found in the local
  docker daemon. The denoted image is packed an OCI artefact set.
`
}