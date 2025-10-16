package ociartifact

import (
	"strings"

	"k8s.io/apimachinery/pkg/util/validation/field"
	"ocm.software/ocm/api/oci/extensions/repositories/docker"
	"ocm.software/ocm/api/oci/grammar"
	"ocm.software/ocm/api/oci/tools/transfer/filters"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/ociartifact"
	"ocm.software/ocm/api/utils/blobaccess"
	ociartifactblob "ocm.software/ocm/api/utils/blobaccess/ociartifact"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs/cpi"
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

func (s *Spec) CreateFilter() ociartifactblob.Option {
	var filter []filters.Filter

	for _, v := range s.Platforms {
		p := strings.Split(v, "/")
		if len(p) == 2 {
			filter = append(filter, filters.Platform(p[0], p[1]))
		}
	}
	if len(filter) > 0 {
		return ociartifactblob.WithFilter(filters.Or(filter...))
	}
	return nil
}

func (s *Spec) GetBlob(ctx inputs.Context, info inputs.InputResourceInfo) (blobaccess.BlobAccess, string, error) {
	filter := s.CreateFilter()
	blob, version, err := ociartifactblob.BlobAccess(s.Path,
		filter,
		ociartifactblob.WithContext(ctx),
		ociartifactblob.WithPrinter(ctx.Printer()),
		ociartifactblob.WithVersion(info.ComponentVersion.GetVersion()),
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
