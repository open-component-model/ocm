package binary

import (
	"k8s.io/apimachinery/pkg/util/validation/field"

	"ocm.software/ocm/api/utils/blobaccess"
	"ocm.software/ocm/api/utils/runtime"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs/cpi"
)

type Spec struct {
	inputs.InputSpecBase `json:",inline"`
	cpi.ProcessSpec      `json:",inline"`

	// Data is plain inline data as byte array
	Data runtime.Binary `json:"data,omitempty"` // json rejects to unmarshal some !string into []byte
}

var _ inputs.InputSpec = (*Spec)(nil)

func New(data []byte, mediatype string, compress bool) *Spec {
	return &Spec{
		InputSpecBase: inputs.InputSpecBase{
			ObjectVersionedType: runtime.ObjectVersionedType{
				Type: TYPE,
			},
		},
		ProcessSpec: cpi.NewProcessSpec(mediatype, compress),
		Data:        (data), // see above
	}
}

func (s *Spec) Validate(fldPath *field.Path, ctx inputs.Context, inputFilePath string) field.ErrorList {
	return nil
}

func (s *Spec) GetBlob(ctx inputs.Context, info inputs.InputResourceInfo) (blobaccess.BlobAccess, string, error) {
	return s.ProcessBlob(ctx, blobaccess.DataAccessForData([]byte(s.Data)), ctx.FileSystem())
}
