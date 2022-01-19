// Copyright 2020 Copyright (c) 2020 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package compdesc

import (
	metav1 "github.com/gardener/ocm/pkg/ocm/compdesc/meta/v1"
	"github.com/gardener/ocm/pkg/runtime"
)

// DefaultComponent applies defaults to a component
func DefaultComponent(component *ComponentDescriptor) {
	if component.RepositoryContexts == nil {
		component.RepositoryContexts = make([]*runtime.UnstructuredTypedObject, 0)
	}
	if component.Sources == nil {
		component.Sources = make([]Source, 0)
	}
	if component.ComponentReferences == nil {
		component.ComponentReferences = make([]ComponentReference, 0)
	}
	if component.Resources == nil {
		component.Resources = make([]Resource, 0)
	}

	if component.Metadata.ConfiguredVersion == "" {
		component.Metadata.ConfiguredVersion = DefaultSchemeVersion
	}
	DefaultResources(component)
}

// DefaultResources defaults a list of resources.
// The version of the component is defaulted for local resources that do not contain a version.
// adds the version as identity if the resource identity would clash otherwise.
func DefaultResources(component *ComponentDescriptor) {
	for i, res := range component.Resources {
		if res.Relation == metav1.LocalRelation && len(res.Version) == 0 {
			component.Resources[i].Version = component.GetVersion()
		}

		id := res.GetIdentity(component.Resources)
		if v, ok := id[SystemIdentityVersion]; ok {
			if res.ExtraIdentity == nil {
				res.ExtraIdentity = metav1.Identity{
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
