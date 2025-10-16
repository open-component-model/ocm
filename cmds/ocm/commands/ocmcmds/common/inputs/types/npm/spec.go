package npm

import (
	"fmt"

	"k8s.io/apimachinery/pkg/util/validation/field"
	"ocm.software/ocm/api/utils/blobaccess"
	"ocm.software/ocm/api/utils/blobaccess/npm"
	"ocm.software/ocm/api/utils/runtime"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs"
)

type Spec struct {
	inputs.InputSpecBase `json:",inline"`
	// Registry is the base URL of the NPM registry
	Registry string `json:"registry"`
	// Package is the name of NPM package
	Package string `json:"package"`
	// Version of the NPM package.
	Version string `json:"version"`
}

var _ inputs.InputSpec = (*Spec)(nil)

func New(registry, pkg, version string) *Spec {
	return &Spec{
		InputSpecBase: inputs.InputSpecBase{
			ObjectVersionedType: runtime.ObjectVersionedType{
				Type: TYPE,
			},
		},
		Registry: registry,
		Package:  pkg,
		Version:  version,
	}
}

func (s *Spec) Validate(fldPath *field.Path, ctx inputs.Context, inputFilePath string) field.ErrorList {
	var allErrs field.ErrorList

	if s.Registry == "" {
		pathField := fldPath.Child("registry")
		allErrs = append(allErrs, field.Invalid(pathField, s.Registry, "no registry"))
	}

	if s.Package == "" {
		pathField := fldPath.Child("package")
		allErrs = append(allErrs, field.Invalid(pathField, s.Package, "no package"))
	}

	return allErrs
}

func (s *Spec) GetBlob(ctx inputs.Context, info inputs.InputResourceInfo) (blobaccess.BlobAccess, string, error) {
	access, err := npm.BlobAccess(s.Registry, s.Package, s.Version, npm.WithDataContext(ctx.OCMContext()))
	if err != nil {
		return nil, "", fmt.Errorf("failed to create access method for npm: %w", err)
	}
	return access, "", err
}
