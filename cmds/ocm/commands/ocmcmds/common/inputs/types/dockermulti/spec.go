package dockermulti

import (
	"fmt"

	"k8s.io/apimachinery/pkg/util/validation/field"

	"ocm.software/ocm/api/oci/extensions/repositories/docker"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/ociartifact"
	"ocm.software/ocm/api/utils/blobaccess"
	"ocm.software/ocm/api/utils/blobaccess/dockermulti"
	"ocm.software/ocm/api/utils/runtime"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs"
	ociartifact2 "ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs/types/ociartifact"
)

type Spec struct {
	inputs.InputSpecBase `json:",inline"`

	// Repository is the repository hint for the index artifact
	Repository string `json:"repository"`
	// Variants holds the list of repository path and tag of the images in the docker daemon
	// used to compose a multi-arch image.
	Variants []string `json:"variants"`
}

var _ inputs.InputSpec = (*Spec)(nil)

func New(pathtags ...string) *Spec {
	return &Spec{
		InputSpecBase: inputs.InputSpecBase{
			ObjectVersionedType: runtime.ObjectVersionedType{
				Type: TYPE,
			},
		},
		Variants: pathtags,
	}
}

func (s *Spec) Validate(fldPath *field.Path, ctx inputs.Context, inputFilePath string) field.ErrorList {
	allErrs := field.ErrorList{}
	allErrs = ociartifact2.ValidateRepository(fldPath.Child("repository"), allErrs, s.Repository)
	variantsField := fldPath.Child("variants")
	if len(s.Variants) == 0 {
		allErrs = append(allErrs, field.Required(variantsField, fmt.Sprintf("variants is required for input of type %q and must has at least one entry", s.GetType())))
	}
	for i, variant := range s.Variants {
		variantField := fldPath.Index(i)
		if variant == "" {
			allErrs = append(allErrs, field.Required(variantField, fmt.Sprintf("non-empty image name is required input of type %q", s.GetType())))
		} else {
			_, _, err := docker.ParseGenericRef(variant)
			if err != nil {
				allErrs = append(allErrs, field.Invalid(variantField, variant, err.Error()))
			}
		}
	}
	return allErrs
}

func (s *Spec) GetBlob(ctx inputs.Context, info inputs.InputResourceInfo) (blobaccess.BlobAccess, string, error) {
	blob, err := dockermulti.BlobAccess(
		dockermulti.WithContext(ctx),
		dockermulti.WithPrinter(ctx.Printer()),
		dockermulti.WithVariants(s.Variants...),
		dockermulti.WithOrigin(info.ComponentVersion),
		dockermulti.WithVersion(info.ComponentVersion.GetVersion()))
	if err != nil {
		return nil, "", err
	}
	return blob, ociartifact.Hint(info.ComponentVersion, info.ElementName, s.Repository, info.ComponentVersion.GetVersion()), nil
}
