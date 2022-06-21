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

// Package jsonv2 provides a normalization which is completely based on the
// abstract (internal) version of the component descriptor and is therefore
// agnostic of the final serialization format. Signatures using this algorithm
// can be transferred among different schema versions, as long as is able to
// handle the complete information using for the normalization.
// Older format might omit some info, therefore the signatures cannot be
// validated for such representations, if the original component descriptor
// has used such parts.
package jsonv2

import (
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	"github.com/open-component-model/ocm/pkg/signing"
)

const Algorithm = compdesc.JsonNormalisationV2

func init() {
	compdesc.Normalizations.Register(Algorithm, normalization{})
}

type normalization struct{}

func (m normalization) Normalize(cd *compdesc.ComponentDescriptor) ([]byte, error) {
	data, err := signing.Normalize(cd, CDExcludes)
	return data, err
}

// CDExcludes describes the fields relevant for Signing
// ATTENTION: if changed, please adapt the HashEqual Functions
// in the generic part, accordingly
var CDExcludes = signing.MapExcludes{
	"meta": nil,
	"component": signing.MapExcludes{
		"repositoryContexts": nil,
		"resources": signing.DynamicArrayExcludes{
			ValueChecker: signing.IgnoreResourcesWithNoneAccess,
			Continue: signing.MapExcludes{
				"access": nil,
				"labels": nil,
				"srcRef": nil,
			},
		},
		"sources": signing.DynamicArrayExcludes{
			ValueChecker: signing.IgnoreResourcesWithNoneAccess,
			Continue: signing.MapExcludes{
				"access": nil,
				"labels": nil,
			},
		},
		"references": signing.ArrayExcludes{
			signing.MapExcludes{
				"labels": nil,
			},
		},
	},
	"signatures": nil,
}
