package ocm

import (
	"k8s.io/apimachinery/pkg/util/validation/field"
	"ocm.software/ocm/api/credentials"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	cpi2 "ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/utils/blobaccess"
	"ocm.software/ocm/api/utils/blobaccess/ocm"
	"ocm.software/ocm/api/utils/runtime"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs"
)

type Spec struct {
	inputs.InputSpecBase `json:",inline"`

	// OCMRepository is the URL of the OCM repository to load the chart from.
	OCMRepository *cpi2.GenericRepositorySpec `json:"ocmRepository,omitempty"`

	// Component if the name of the root component used to lookup the resource.
	Component string `json:"component,omitempty"`

	// Version is the version og the root component.
	Version string `json:"version,omitempty,"`

	ResourceRef metav1.ResourceReference `json:"resourceRef"`
}

var _ inputs.InputSpec = (*Spec)(nil)

func New(comp, vers string, repo cpi2.RepositorySpec, id metav1.Identity, path ...metav1.Identity) (inputs.InputSpec, error) {
	spec, err := cpi2.ToGenericRepositorySpec(repo)
	if err != nil {
		return nil, err
	}
	return &Spec{
		InputSpecBase: inputs.InputSpecBase{
			ObjectVersionedType: runtime.ObjectVersionedType{
				Type: TYPE,
			},
		},
		OCMRepository: spec,
		Component:     comp,
		Version:       vers,
		ResourceRef:   metav1.NewNestedResourceRef(id, path),
	}, nil
}

func (s *Spec) Validate(fldPath *field.Path, ctx inputs.Context, inputFilePath string) field.ErrorList {
	var allErrs field.ErrorList

	err := s.OCMRepository.Validate(ctx.OCMContext(), nil, credentials.StringUsageContext(s.Component))
	if err != nil {
		data, _ := s.OCMRepository.MarshalJSON()
		pathField := fldPath.Child("ocmRepository")
		allErrs = append(allErrs, field.Invalid(pathField, string(data), err.Error()))
	}
	if s.Component == "" {
		pathField := fldPath.Child("component")
		allErrs = append(allErrs, field.Invalid(pathField, s.Component, "no component name"))
	}
	if s.Version == "" {
		pathField := fldPath.Child("version")
		allErrs = append(allErrs, field.Invalid(pathField, s.Version, "no component version"))
	}

	return allErrs
}

func (s *Spec) GetBlob(ctx inputs.Context, info inputs.InputResourceInfo) (blobaccess.BlobAccess, string, error) {
	b, err := ocm.BlobAccess(ocm.ByRepositorySpecAndName(ctx.OCMContext(), s.OCMRepository, s.Component, s.Version), ocm.ByResourceRef(s.ResourceRef))
	return b, "", err
}
