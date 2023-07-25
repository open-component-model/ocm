// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package binary

import (
	"k8s.io/apimachinery/pkg/util/validation/field"

	"github.com/open-component-model/ocm/v2/cmds/ocm/commands/ocmcmds/common/inputs"
	"github.com/open-component-model/ocm/v2/cmds/ocm/commands/ocmcmds/common/inputs/cpi"
	"github.com/open-component-model/ocm/v2/pkg/common/accessio"
	"github.com/open-component-model/ocm/v2/pkg/runtime"
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

func (s *Spec) GetBlob(ctx inputs.Context, info inputs.InputResourceInfo) (accessio.TemporaryBlobAccess, string, error) {
	return s.ProcessBlob(ctx, accessio.DataAccessForBytes([]byte(s.Data)), ctx.FileSystem())
}
