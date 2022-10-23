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

package add

import (
	"encoding/json"

	ocmcomm "github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common"
	"github.com/open-component-model/ocm/pkg/cobrautils/flagsets"
)

type ReferenceResourceSpecificationProvider struct {
	*ocmcomm.ResourceMetaDataSpecificationsProvider
}

var _ ocmcomm.ResourceSpecificationsProvider = (*ReferenceResourceSpecificationProvider)(nil)
var _ ocmcomm.ResourceSpecifications = (*ReferenceResourceSpecificationProvider)(nil)

func NewReferenceSpecificatonProvider() ocmcomm.ResourceSpecificationsProvider {
	a := &ReferenceResourceSpecificationProvider{
		ResourceMetaDataSpecificationsProvider: ocmcomm.NewResourceMetaDataSpecificationsProvider("reference", addMeta,
			flagsets.NewStringOptionType("component", "component name"),
		),
	}
	return a
}

func addMeta(opts flagsets.ConfigOptions, config flagsets.Config) error {
	flagsets.AddFieldByOption(opts, "component", config, "componentName")
	return nil
}

func (a *ReferenceResourceSpecificationProvider) Description() string {
	return a.ResourceMetaDataSpecificationsProvider.Description() + `
The component name can be specified with the option <code>--component</code>. 
Therefore, basic references not requiring any additional labels or extra
identities can just be specified by those simple value options without the need
for the YAML option.
`
}

func (a *ReferenceResourceSpecificationProvider) Get() (string, error) {
	data, err := a.ParsedMeta()
	if err != nil {
		return "", err
	}

	r, err := json.Marshal(data)
	return string(r), nil
}

func (a *ReferenceResourceSpecificationProvider) Resources() ([]ocmcomm.ResourceSpecifications, error) {
	if !a.IsSpecified() {
		return nil, nil
	}
	return []ocmcomm.ResourceSpecifications{a}, nil
}
