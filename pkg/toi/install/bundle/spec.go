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

package bundle

import (
	"encoding/json"

	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/toi/install"
)

type BundleSpecification struct {
	install.CredentialsRequest `json:",inline"`
	Template                   json.RawMessage            `json:"configTemplate,omitempty"`
	Libraries                  []metav1.ResourceReference `json:"templateLibraries,omitempty"`
	Scheme                     json.RawMessage            `json:"configScheme,omitempty"`

	Actions []string          `json:"actions,omitempty"`
	Outputs map[string]string `json:"outputs,omitempty"`
}

type InstallationSpecification struct {
	ResourceRef *metav1.ResourceReference `json:"resourceRef,omitempty"`
	Image       *install.Image            `json:"image,omitempty"`

	Actions  map[string]string `json:"actions,omitempty"`
	Required map[string]string `json:"required,omitempty"`
}

type InstallationValues struct {
	install.Credentials `json:",inline"`
	Settings            json.RawMessage `json:"values,omitempty"`
}
