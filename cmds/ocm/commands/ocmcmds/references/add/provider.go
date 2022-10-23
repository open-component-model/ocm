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
	"fmt"

	"github.com/spf13/pflag"

	ocmcomm "github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common"
	"github.com/open-component-model/ocm/pkg/errors"
)

type ReferenceResourceSpecificationProvider struct {
	*ocmcomm.ResourceMetaDataSpecificationsProvider
	component string
}

var _ ocmcomm.ResourceSpecificationsProvider = (*ReferenceResourceSpecificationProvider)(nil)
var _ ocmcomm.ResourceSpecifications = (*ReferenceResourceSpecificationProvider)(nil)

func NewReferenceSpecificatonProvider() ocmcomm.ResourceSpecificationsProvider {
	return &ReferenceResourceSpecificationProvider{
		ResourceMetaDataSpecificationsProvider: ocmcomm.NewResourceMetaDataSpecificationsProvider("reference"),
	}
}

func (a *ReferenceResourceSpecificationProvider) Description() string {
	return a.ResourceMetaDataSpecificationsProvider.Description() + `
The component name can be specified with the option <code>--component</code>. 
Therefore, basic references not requiring any additional labels or extra
identities can just be specified by those simple value options without the need
for the YAML option.
`
}

func (a *ReferenceResourceSpecificationProvider) AddFlags(fs *pflag.FlagSet) {
	a.ResourceMetaDataSpecificationsProvider.AddFlags(fs)
	fs.StringVarP(&a.component, "component", "", "", "component name")
}

func (a *ReferenceResourceSpecificationProvider) IsSpecified() bool {
	return a.ResourceMetaDataSpecificationsProvider.IsSpecified() || a.component != ""
}

func (a *ReferenceResourceSpecificationProvider) Complete() error {
	if !a.IsSpecified() {
		return nil
	}
	generic := a.ResourceMetaDataSpecificationsProvider.IsSpecified()
	if generic {
		if err := a.ResourceMetaDataSpecificationsProvider.Complete(); err != nil {
			return err
		}
	} else {
		if a.component == "" {
			return fmt.Errorf("--component is required")
		}
	}
	return nil
}

func (a *ReferenceResourceSpecificationProvider) Resources() ([]ocmcomm.ResourceSpecifications, error) {
	if !a.IsSpecified() {
		return nil, nil
	}
	return []ocmcomm.ResourceSpecifications{a}, nil
}

func (a *ReferenceResourceSpecificationProvider) Get() (string, error) {
	data, err := a.ParsedMeta()
	if err != nil {
		return "", err
	}

	if a.component != "" {
		data["componentName"] = a.component
	}

	r, err := json.Marshal(data)
	if err != nil {
		return "", errors.Wrapf(err, "cannot marshal %s", a.Origin())
	}
	return string(r), nil
}
