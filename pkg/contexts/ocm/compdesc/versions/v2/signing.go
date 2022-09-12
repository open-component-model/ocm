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

package compdesc

import (
	"fmt"

	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	"github.com/open-component-model/ocm/pkg/signing"
)

// CDExcludes describes the fields relevant for Signing
// ATTENTION: if changed, please adapt the HashEqual Functions
// in the generic part, accordingly.
var CDExcludes = signing.MapExcludes{
	"component": signing.MapExcludes{
		"labels": signing.ExcludeEmpty{signing.DynamicArrayExcludes{
			ValueChecker: signing.IgnoreLabelsWithoutSignature,
			Continue:     signing.NoExcludes{},
		}},
		"repositoryContexts": nil,
		"resources": signing.DynamicArrayExcludes{
			ValueChecker: signing.IgnoreResourcesWithNoneAccess,
			Continue: signing.MapExcludes{
				"access": nil,
				"srcRef": nil,
				"labels": signing.ExcludeEmpty{signing.DynamicArrayExcludes{
					ValueChecker: signing.IgnoreLabelsWithoutSignature,
					Continue:     signing.NoExcludes{},
				}},
			},
		},
		"sources": signing.DynamicArrayExcludes{
			ValueChecker: signing.IgnoreResourcesWithNoneAccess,
			Continue: signing.MapExcludes{
				"access": nil,
				"labels": signing.ExcludeEmpty{signing.DynamicArrayExcludes{
					ValueChecker: signing.IgnoreLabelsWithoutSignature,
					Continue:     signing.NoExcludes{},
				}},
			},
		},
		"references": signing.ArrayExcludes{
			signing.MapExcludes{
				"labels": signing.ExcludeEmpty{signing.DynamicArrayExcludes{
					ValueChecker: signing.IgnoreLabelsWithoutSignature,
					Continue:     signing.NoExcludes{},
				}},
			},
		},
	},
	"signatures": nil,
}

func (cd *ComponentDescriptor) Normalize(normAlgo string) ([]byte, error) {
	if normAlgo != compdesc.JsonNormalisationV1 {
		return nil, fmt.Errorf("unsupported cd normalization %q", normAlgo)
	}
	data, err := signing.Normalize(cd, CDExcludes)
	// fmt.Printf("**** normalized:\n %s\n", string(data))
	return data, err
}
