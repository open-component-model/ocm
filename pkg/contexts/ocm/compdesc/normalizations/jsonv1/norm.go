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

// Package jsonv1 provides a normalization which uses schema specific
// normalizations.
// It creates the requested schema for the component descriptor
// and just forwards the normalization to this version.
package jsonv1

import (
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	"github.com/open-component-model/ocm/pkg/errors"
)

const Algorithm = compdesc.JsonNormalisationV1

func init() {
	compdesc.Normalizations.Register(Algorithm, normalization{})
}

type normalization struct{}

func (m normalization) Normalize(cd *compdesc.ComponentDescriptor) ([]byte, error) {
	cv := compdesc.DefaultSchemes[cd.SchemaVersion()]
	if cv == nil {
		if cv == nil {
			return nil, errors.ErrNotSupported(errors.KIND_SCHEMAVERSION, cd.SchemaVersion())
		}
	}
	v, err := cv.ConvertFrom(cd)
	if err != nil {
		return nil, err
	}
	return v.Normalize(Algorithm)
}
