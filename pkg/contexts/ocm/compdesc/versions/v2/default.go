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
	v1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/runtime"
)

// Default applies defaults to a component
func (cd *ComponentDescriptor) Default() error {
	if cd.RepositoryContexts == nil {
		cd.RepositoryContexts = make([]*runtime.UnstructuredTypedObject, 0)
	}
	if cd.Sources == nil {
		cd.Sources = make([]Source, 0)
	}
	if cd.ComponentReferences == nil {
		cd.ComponentReferences = make([]ComponentReference, 0)
	}
	if cd.Resources == nil {
		cd.Resources = make([]Resource, 0)
	}

	DefaultResources(cd)
	return nil
}

// DefaultResources defaults a list of resources.
// The version of the component is defaulted for local resources that do not contain a version.
// adds the version as identity if the resource identity would clash otherwise.
func DefaultResources(component *ComponentDescriptor) {
	for i, res := range component.Resources {
		if res.Relation == v1.LocalRelation && len(res.Version) == 0 {
			component.Resources[i].Version = component.GetVersion()
		}

		id := res.GetIdentity(component.Resources)
		if v, ok := id[SystemIdentityVersion]; ok {
			if res.ExtraIdentity == nil {
				res.ExtraIdentity = v1.Identity{
					SystemIdentityVersion: v,
				}
			} else {
				if _, ok := res.ExtraIdentity[SystemIdentityVersion]; !ok {
					res.ExtraIdentity[SystemIdentityVersion] = v
				}
			}
		}
	}
}
