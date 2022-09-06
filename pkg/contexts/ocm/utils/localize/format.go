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

package localize

import (
	"encoding/json"
	"fmt"

	v1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/runtime"
)

// Definition of inbound substitution requests.
// - Localizations  image locations substitution resolved using a component version
// - Configurations configuration substitution resolved by provided value data.
//
// Such requests can be given to merge externally provided data into
// some filesystem template.
// The evaluation of such requests results in a list
// of resolved substitution requests that can be applied without
// further value context to a filesystem structure.

// ImageMapping describes a dedicated substitution of parts
// of container image names based on a relative OCM resource reference.
type ImageMapping struct {
	// The resource reference used to resolve the substitution
	v1.ResourceReference `json:",inline"`

	// The optional variants for the value determination

	// Path in target to substitute by the image tag/digest
	Tag string `json:"tag,omitempty"`
	// Path in target to substitute the image repository
	Repository string `json:"repository,omitempty"`
	// Path in target to substitute the complete image
	Image string `json:"image,omitempty"`
}

// Localization is a request to substitute an image location.
// The specification describes substitution targets given by the file path and
// the YAML/JSON value paths of the elements in this file.
// The substitution value is calculated
// from the access specification of the given resource provided by the actual
// component version.
type Localization struct {
	// The optional but unique(!) name of the mapping to support referencing mapping entries
	Name string `json:"name,omitempty"`
	// The path of the file for the substitution
	FilePath string `json:"file"`
	// The image mapping request
	ImageMapping `json:",inline"`
}

// Configuration is a request to substitute a configuration value.
// The specification describes substitution targets given by the file path and
// the YAML/JSON value paths of the elements in this file.
// The substitution value is calculated
// by the value expression (spiff) based on given config data.
// It has the same structure as Substitution, but is a request based
// on external configuration data, while a Substitution describes a fixed target
// value.
type Configuration Substitution

// Here comes the structure used for resolved execution requests.
// They can be applied to a filesystem content without further external information.
// It basically has the same structure as the configuration request, but
// the given value is just the target value without any further interpretation.
// This way configuration requests could just provide dedicated values, also

// Substitution is a request to substitute the YAML/JSON
// element given by the value path in the given file path by the given
// direct value.
type Substitution struct {
	// The optional but unique(!) name of the mapping to support referencing mapping entries
	Name string `json:"name,omitempty"`
	// The path of the file for the substitution
	FilePath string `json:"file"`
	// The target path for the value substitution
	ValuePath string `json:"path"`
	// The value to set
	Value json.RawMessage `json:"value"`
}

func (s *Substitution) GetValue() (interface{}, error) {
	var value interface{}
	err := runtime.DefaultYAMLEncoding.Unmarshal(s.Value, &value)
	return value, err
}

type Substitutions []Substitution

func (s *Substitutions) Add(name, file, path string, value interface{}) error {
	var (
		v   []byte
		err error
	)

	if value != nil {
		v, err = runtime.DefaultJSONEncoding.Marshal(value)
		if err != nil {
			return fmt.Errorf("cannot marshal substitution value: %w", err)
		}
	}
	*s = append(*s, Substitution{
		Name:      name,
		FilePath:  file,
		ValuePath: path,
		Value:     v,
	})
	return nil
}

// InstantiationRules bundle the localization of a filesystem resource
// covering image localization and applying instance configuration.
type InstantiationRules struct {
	Template          v1.ResourceReference   `json:"templateResource,omitempty"`
	LocalizationRules []Localization         `json:"localizationRules,omitempty"`
	ConfigRules       []Configuration        `json:"configRules,omitempty"`
	ConfigScheme      json.RawMessage        `json:"configScheme,omitempty"`
	ConfigTemplate    json.RawMessage        `json:"configTemplate,omitempty"`
	ConfigLibraries   []v1.ResourceReference `json:"configLibraries,omitempty"`
}
