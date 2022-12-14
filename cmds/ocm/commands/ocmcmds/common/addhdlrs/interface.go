// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package addhdlrs

import (
	"encoding/json"
	"fmt"

	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/inputs"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/runtime"
)

type ResourceInput struct {
	Access *runtime.UnstructuredTypedObject `json:"access"`
	// Input  *inputs.BlobInput                `json:"input,omitempty"`
	Input *inputs.GenericInputSpec `json:"input,omitempty"`
}

type ElementSpecHandler interface {
	Key() string
	RequireInputs() bool
	Decode(data []byte) (ElementSpec, error)
}

type ElementSource interface {
	Origin() SourceInfo
	Get() (string, error)
}

type SourceInfo interface {
	Origin() string
	Id() string

	String() string
	Sub(indices ...interface{}) SourceInfo
}

type sourceInfo struct {
	origin  string
	indices []interface{}
}

func NewSourceInfo(origin string) SourceInfo {
	return &sourceInfo{origin: origin}
}

func (s *sourceInfo) Sub(indices ...interface{}) SourceInfo {
	if len(indices) == 0 {
		return s
	}
	return &sourceInfo{
		origin:  s.origin,
		indices: append(s.indices, indices...),
	}
}

func (s *sourceInfo) String() string {
	return s.Id()
}

func (s *sourceInfo) Origin() string {
	return s.origin
}

func (s *sourceInfo) Id() string {
	id := s.origin
	for _, i := range s.indices {
		id += fmt.Sprintf("[%v]", i)
	}
	return id
}

type ElementSpec interface {
	GetName() string
	Info() string
	Validate(ctx clictx.Context, input *ResourceInput) error
}

type Element interface {
	Source() SourceInfo
	Spec() ElementSpec
	Type() string
	Data() []byte
	Input() *ResourceInput
}

func NewElement(spec ElementSpec, input *ResourceInput, src SourceInfo, data []byte, indices ...interface{}) Element {
	return &element{
		source: src.Sub(indices...),
		spec:   spec,
		data:   data,
		input:  input,
	}
}

type element struct {
	source SourceInfo
	spec   ElementSpec
	data   []byte
	input  *ResourceInput
}

func (r *element) Source() SourceInfo {
	return r.source
}

func (r *element) Spec() ElementSpec {
	return r.spec
}

func (r *element) Data() []byte {
	return r.data
}

func (r *element) Input() *ResourceInput {
	return r.input
}

func (r *element) Type() string {
	if c, ok := r.spec.(interface{ GetType() string }); ok {
		return c.GetType()
	}
	return ""
}

func MapSpecsToElems[T ElementSpec](ctx clictx.Context, ictx inputs.Context, si SourceInfo, specs []T, h ElementSpecHandler) ([]Element, error) {
	var result []Element
	for i, e := range specs {
		data, err := json.Marshal(e)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot marshal element %d", i+1)
		}
		elem, err := DetermineElementForData(ctx, ictx, si.Sub(i), data, h)
		if err != nil {
			return nil, errors.Wrapf(err, "entry %d", i+1)
		}
		result = append(result, elem)
	}
	return result, nil
}
