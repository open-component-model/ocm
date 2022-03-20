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

package core

import (
	"fmt"

	"github.com/gardener/ocm/pkg/oci/repositories/ocireg"
)

// UniformRepositorySpec is is generic specification of the repository
// for handling as part of standard references
type UniformRepositorySpec struct {
	// Type
	Type string `json:"type,omitempty"`
	// Host is the hostname of an ocm ref.
	Host string `json:"host,omitempty"`
	// SubPath is the sub path spec used to host component versions
	SubPath string `json:"subPath,omitempty"`

	// Info contains Type specifif information
	Info map[string]string `json:"info,omitempty"`
}

func (u *UniformRepositorySpec) String() string {
	t := u.Type
	if t != "" && t != ocireg.OCIRegistryRepositoryType {
		t = t + "::"
	}
	s := u.SubPath
	if s != "" {
		s = "/" + s
	}
	return fmt.Sprintf("%s%s%s", t, u.Host, s)
}
