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

package ociimage

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/util/validation/field"

	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/inputs"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/inputs/cpi"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/inputs/options"
	"github.com/open-component-model/ocm/pkg/clisupport"
	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/contexts/oci"
	"github.com/open-component-model/ocm/pkg/contexts/oci/grammar"
	"github.com/open-component-model/ocm/pkg/contexts/oci/repositories/artefactset"
	"github.com/open-component-model/ocm/pkg/contexts/oci/repositories/docker"
)

func ConfigHandler() clisupport.ConfigOptionTypeSetHandler {
	return cpi.NewMediaFileSpecOptionType(TYPE, AddConfig,
		options.PathOption, options.HintOption)
}

type Spec struct {
	// PathSpec holds the repository path and tag of the image in the docker daemon
	cpi.PathSpec
	// Repository is the repository hint for the index artefact
	Repository string `json:"repository,omitempty"`
}

var _ inputs.InputSpec = (*Spec)(nil)

func New(pathtag string) *Spec {
	return &Spec{
		PathSpec: cpi.NewPathSpec(TYPE, pathtag),
	}
}

func (s *Spec) Validate(fldPath *field.Path, ctx inputs.Context, inputFilePath string) field.ErrorList {
	allErrs := s.PathSpec.Validate(fldPath, ctx, inputFilePath)
	allErrs = ValidateRepository(fldPath.Child("repository"), allErrs, s.Repository)
	if s.Path != "" {
		pathField := fldPath.Child("path")
		_, _, err := docker.ParseGenericRef(s.Path)
		if err != nil {
			allErrs = append(allErrs, field.Invalid(pathField, s.Path, err.Error()))
		}
	}
	return allErrs
}

func (s *Spec) GetBlob(ctx inputs.Context, nv common.NameVersion, inputFilePath string) (accessio.TemporaryBlobAccess, string, error) {
	ctx.Printf("image %s\n", s.Path)
	ref, err := oci.ParseRef(s.Path)
	if err != nil {
		return nil, "", err
	}

	spec, err := ctx.OCIContext().MapUniformRepositorySpec(&ref.UniformRepositorySpec)
	if err != nil {
		return nil, "", err
	}

	repo, err := ctx.OCIContext().RepositoryForSpec(spec)
	if err != nil {
		return nil, "", err
	}
	ns, err := repo.LookupNamespace(ref.Repository)
	if err != nil {
		return nil, "", err
	}

	version := ref.Version()
	if version == "" || version == "latest" {
		version = nv.GetVersion()
	}
	blob, err := artefactset.SynthesizeArtefactBlob(ns, version)
	if err != nil {
		return nil, "", err
	}
	return blob, Hint(nv, ref.Repository, s.Repository, version), nil
}

func Hint(nv common.NameVersion, locator, repo, version string) string {
	fmt.Printf("locator: %s, repo: %s, version %s\n", locator, repo, version)
	repository := fmt.Sprintf("%s/%s", nv.GetName(), locator)
	if repo != "" {
		if strings.HasPrefix(repo, grammar.RepositorySeparator) {
			repository = repo[1:]
		} else {
			repository = fmt.Sprintf("%s/%s", nv.GetName(), repo)
		}
	}
	return fmt.Sprintf("%s:%s", repository, version)
}

func ValidateRepository(fldPath *field.Path, allErrs field.ErrorList, repo string) field.ErrorList {
	if repo == "" {
		return allErrs
	}
	if strings.Contains(repo, grammar.DigestSeparator) || strings.Contains(repo, grammar.TagSeparator) {
		return append(allErrs, field.Invalid(fldPath, repo, "unexpected digest or tag"))
	}
	if !grammar.AnchoredRepositoryRegexp.MatchString(repo) {
		return append(allErrs, field.Invalid(fldPath, repo, "no repository name"))
	}
	return allErrs
}

func AddConfig(opts clisupport.ConfigOptions, config clisupport.Config) error {
	if err := cpi.AddPathSpecConfig(opts, config); err != nil {
		return err
	}
	if v, ok := opts.GetValue(options.HintOption.Name()); ok {
		config["repository"] = v
	}
	return nil
}
