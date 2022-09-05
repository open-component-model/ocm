// Copyright 2022 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package spiff

import (
	"encoding/json"
	"fmt"

	"github.com/mandelsoft/spiff/features"
	"github.com/mandelsoft/spiff/spiffing"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/modern-go/reflect2"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/errors"
)

type Request struct {
	Template   spiffing.Source
	Stubs      []spiffing.Source
	ValuesNode string
	Values     interface{}
	FileSystem vfs.FileSystem
}

func (r Request) GetValues() (map[string]interface{}, error) {
	if reflect2.IsNil(r.Values) {
		return nil, nil
	}

	data, err := json.Marshal(r.Values)
	if err != nil {
		return nil, errors.ErrInvalidWrap(err, "values", fmt.Sprintf("%T", r.Values))
	}

	var values interface{}
	err = json.Unmarshal(data, &values)
	if err != nil {
		return nil, errors.ErrInvalidWrap(err, "values", fmt.Sprintf("%T", r.Values))
	}
	if r.ValuesNode != "" {
		return map[string]interface{}{r.ValuesNode: values}, nil
	}
	if v, ok := values.(map[string]interface{}); ok {
		return v, nil
	}
	return nil, errors.ErrInvalid("values", fmt.Sprintf("%T", values))
}

func (r *Request) GetSpiff() (spiffing.Spiff, error) {
	spiff := spiffing.New().WithFeatures(features.CONTROL, features.INTERPOLATION).WithFileSystem(accessio.FileSystem(r.FileSystem))
	values, err := r.GetValues()
	if err != nil {
		return nil, err
	}
	if values != nil {
		spiff, err = spiff.WithValues(values)
	}
	if err != nil {
		return nil, err
	}
	return spiff, nil
}

func Cascade(req *Request) ([]byte, error) {
	if req.Template == nil {
		return nil, nil
	}
	spiff, err := req.GetSpiff()
	if err != nil {
		return nil, err
	}
	stubs := []spiffing.Node{}

	data, err := req.Template.Data()
	if err != nil {
		return nil, errors.ErrInvalidWrap(err, "template", req.Template.Name())
	}
	templ, err := spiff.Unmarshal("template "+req.Template.Name(), data)
	if err != nil {
		return nil, errors.Wrapf(err, "template: %s", req.Template.Name())
	}

	for i, s := range req.Stubs {
		data, err := s.Data()
		if err != nil {
			return nil, errors.ErrInvalidWrap(err, "stub", s.Name())
		}
		stub, err := spiff.Unmarshal(s.Name(), data)
		if err != nil {
			return nil, errors.Wrapf(err, "stub %d (%s)", i+1, s.Name())
		}
		stubs = append(stubs, stub)
	}

	node, err := spiff.Cascade(templ, stubs)
	if err != nil {
		return nil, errors.Wrapf(err, "processing template %s", req.Template.Name())
	}
	return spiff.Marshal(node)
}

func CascadeWith(opts ...Option) ([]byte, error) {
	req, err := GetRequest(opts...)
	if err != nil {
		return nil, err
	}
	return Cascade(req)
}
