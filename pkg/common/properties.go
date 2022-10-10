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

package common

import (
	"encoding/json"

	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/sets"
)

// Properties describes a set of name/value pairs.
type Properties map[string]string

// Digest returns the object digest of an Property set.
func (p Properties) Digest() []byte {
	data, err := json.Marshal(p)
	if err != nil {
		logrus.Error(err)
	}
	return data
}

func (p Properties) SetNonEmptyValue(name, value string) {
	if value != "" {
		p[name] = value
	}
}

// Equals compares two identities.
func (p Properties) Equals(o Properties) bool {
	if len(p) != len(o) {
		return false
	}

	for k, v := range p {
		if v2, ok := o[k]; !ok || v != v2 {
			return false
		}
	}
	return true
}

// Match implements the selector interface.
func (p Properties) Match(obj map[string]string) (bool, error) {
	for k, v := range p {
		if obj[k] != v {
			return false, nil
		}
	}
	return true, nil
}

// Names returns the set of property names.
func (c Properties) Names() sets.String {
	return sets.StringKeySet(c)
}

// String returns a string representation.
func (c Properties) String() string {
	if c == nil {
		return "<none>"
	}
	//nolint: errchkjson // just a string map
	d, _ := json.Marshal(c)
	return string(d)
}

// Copy copies identity.
func (p Properties) Copy() Properties {
	if p == nil {
		return nil
	}
	n := Properties{}
	for k, v := range p {
		n[k] = v
	}
	return n
}
