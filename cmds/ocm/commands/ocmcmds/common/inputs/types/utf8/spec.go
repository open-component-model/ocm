// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package utf8

import (
	"k8s.io/apimachinery/pkg/util/validation/field"

	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/inputs"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/inputs/cpi"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/runtime"
)

type Spec struct {
	runtime.ObjectVersionedType `json:",inline"`
	cpi.ProcessSpec             `json:",inline"`

	// Text is an utf8 string
	Text string `json:"text,omitempty"`
}

var _ inputs.InputSpec = (*Spec)(nil)

func New(text string, mediatype string, compress bool) *Spec {
	return &Spec{
		ObjectVersionedType: runtime.ObjectVersionedType{
			Type: TYPE,
		},
		ProcessSpec: cpi.NewProcessSpec(mediatype, compress),
		Text:        text,
	}
}

func (s *Spec) Validate(fldPath *field.Path, ctx inputs.Context, inputFilePath string) field.ErrorList {
	return nil
}

func (s *Spec) GetBlob(ctx inputs.Context, info inputs.InputResourceInfo) (accessio.TemporaryBlobAccess, string, error) {
	return s.ProcessBlob(ctx, accessio.DataAccessForBytes([]byte(s.Text)), ctx.FileSystem())
}
