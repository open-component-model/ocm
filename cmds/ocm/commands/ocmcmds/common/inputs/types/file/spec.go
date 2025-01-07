package file

import (
	"k8s.io/apimachinery/pkg/util/validation/field"

	"ocm.software/ocm/api/ocm/refhints"
	"ocm.software/ocm/api/utils/blobaccess"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs/cpi"
)

type Spec struct {
	cpi.MediaFileSpec `json:",inline"`
	ReferenceHints    refhints.DefaultReferenceHints `json:"referenceHints,omitempty"`
}

var _ inputs.InputSpec = (*Spec)(nil)

func New(path, mediatype string, compress bool) *Spec {
	return &Spec{
		MediaFileSpec: cpi.NewMediaFileSpec(TYPE, path, mediatype, compress),
	}
}

func (s *Spec) Validate(fldPath *field.Path, ctx inputs.Context, inputFilePath string) field.ErrorList {
	return (&FileProcessSpec{s.MediaFileSpec, s.ReferenceHints, nil}).Validate(fldPath, ctx, inputFilePath)
}

func (s *Spec) GetBlob(ctx inputs.Context, info inputs.InputResourceInfo) (blobaccess.BlobAccess, []refhints.ReferenceHint, error) {
	return (&FileProcessSpec{s.MediaFileSpec, refhints.AsImplicit(s.ReferenceHints), nil}).GetBlob(ctx, info)
}
