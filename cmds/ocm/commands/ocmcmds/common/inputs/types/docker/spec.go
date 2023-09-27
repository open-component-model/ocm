// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package docker

import (
	"github.com/open-component-model/ocm/pkg/contexts/oci"
	"github.com/open-component-model/ocm/pkg/contexts/oci/annotations"
	"k8s.io/apimachinery/pkg/util/validation/field"

	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/inputs"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/inputs/cpi"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/inputs/types/ociimage"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/contexts/oci/repositories/artifactset"
	"github.com/open-component-model/ocm/pkg/contexts/oci/repositories/docker"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/ociartifact"
)

type Spec struct {
	// PathSpec holds the repository path and tag of the image in the docker daemon
	cpi.PathSpec
	// Repository is the repository hint for the index artifact
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
	allErrs = ociimage.ValidateRepository(fldPath.Child("repository"), allErrs, s.Repository)

	if s.Path != "" {
		pathField := fldPath.Child("path")
		_, _, err := docker.ParseGenericRef(s.Path)
		if err != nil {
			allErrs = append(allErrs, field.Invalid(pathField, s.Path, err.Error()))
		}
	}
	return allErrs
}

func (s *Spec) GetBlob(ctx inputs.Context, info inputs.InputResourceInfo) (accessio.TemporaryBlobAccess, string, error) {
	ctx.Printf("image %s\n", s.Path)
	locator, version, err := docker.ParseGenericRef(s.Path)
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

	if version == "" || version == "latest" {
		version = info.ComponentVersion.GetVersion()
	}
	blob, err := artifactset.SynthesizeArtifactBlob(ns, version,
		func(art oci.ArtifactAccess) error {
			art.Artifact().SetAnnotation(annotations.COMPVERS_ANNOTATION, info.ComponentVersion.String())
			return nil
		},
	)
	if err != nil {
		return nil, "", err
	}
	return blob, ociartifact.Hint(info.ComponentVersion, locator, s.Repository, version), nil
}
