// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package ociartifact

import (
	"strings"

	"k8s.io/apimachinery/pkg/util/validation/field"

	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/inputs"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/inputs/cpi"
	"github.com/open-component-model/ocm/pkg/blobaccess"
	ociartifact2 "github.com/open-component-model/ocm/pkg/blobaccess/ociartifact"
	"github.com/open-component-model/ocm/pkg/contexts/oci/grammar"
	"github.com/open-component-model/ocm/pkg/contexts/oci/repositories/docker"
	"github.com/open-component-model/ocm/pkg/contexts/oci/transfer/filters"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/ociartifact"
)

type Spec struct {
	// PathSpec holds the repository path and tag of the image in the docker daemon
	cpi.PathSpec
	// Repository is the repository hint for the index artifact
	Repository string `json:"repository,omitempty"`
	// Platforms provides filters for operating system/architecture.
	// Syntax [OS]/[Architecture]
	Platforms []string `json:"platforms,omitempty"`
}

var _ inputs.InputSpec = (*Spec)(nil)

func New(pathtag string, platforms ...string) *Spec {
	return &Spec{
		PathSpec:  cpi.NewPathSpec(TYPE, pathtag),
		Platforms: platforms,
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

func (s *Spec) CreateFilter() ociartifact2.Option {
	var filter []filters.Filter

	for _, v := range s.Platforms {
		p := strings.Split(v, "/")
		if len(p) == 2 {
			filter = append(filter, filters.Platform(p[0], p[1]))
		}
	}
	if len(filter) > 0 {
		return ociartifact2.WithFilter(filters.Or(filter...))
	}
	return nil
}

func (s *Spec) GetBlob(ctx inputs.Context, info inputs.InputResourceInfo) (blobaccess.BlobAccess, string, error) {
	filter := s.CreateFilter()
	blob, version, err := ociartifact2.BlobAccessForOCIArtifact(s.Path,
		filter,
		ociartifact2.WithContext(ctx),
		ociartifact2.WithPrinter(ctx.Printer()),
		ociartifact2.WithVersion(info.ComponentVersion.GetVersion()),
	)
	if err != nil {
		return nil, "", err
	}
	return blob, ociartifact.Hint(info.ComponentVersion, info.ElementName, s.Repository, version), nil
}

func ValidateRepository(fldPath *field.Path, allErrs field.ErrorList, repo string) field.ErrorList {
	if repo == "" {
		return allErrs
	}
	if strings.Contains(repo, grammar.DigestSeparator) || strings.Contains(repo, grammar.TagSeparator) {
		return append(allErrs, field.Invalid(fldPath, repo, "unexpected digest or tag"))
	}

	if !grammar.AnchoredRepositoryRegexp.MatchString(strings.TrimPrefix(repo, "/")) {
		return append(allErrs, field.Invalid(fldPath, repo, "no repository name"))
	}
	return allErrs
}
