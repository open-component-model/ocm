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

package app

import (
	"encoding/json"

	v1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/errors"
)

type Config struct {
	Chart           v1.ResourceReference `json:"chart,omitempty"`
	Release         string               `json:"release,omitempty"`
	Namespace       string               `json:"namespace,omitempty"`
	CreateNamespace bool                 `json:"createNamespace,omitempty"`
	ImageMapping    []ImageMapping       `json:"imageMapping"`
	Values          json.RawMessage      `json:"values,omitempty"`
	KubeConfigName  string               `json:"kubeConfigName,omitempty"`
}

type ImageMapping struct {
	v1.ResourceReference
	Tag        string `json:"tag,omitempty"`
	Repository string `json:"repository,omitempty"`
	Image      string `json:"image,omitempty"`
}

func (c *Config) GetValues() (map[string]interface{}, error) {
	if len(c.Values) == 0 {
		return map[string]interface{}{}, nil
	}
	var result map[string]interface{}
	err := json.Unmarshal(c.Values, &result)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot unmarshal values from bootstrap config")
	}
	return result, nil
}
