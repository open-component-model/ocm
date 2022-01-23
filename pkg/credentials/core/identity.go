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

package core

import (
	"encoding/json"
)

// ConsumerIdentity describes the identity of a credential consumer.
type ConsumerIdentity map[string]string

// IdentityByURL return a simple url identity
func IdentityByURL(url string) ConsumerIdentity {
	return ConsumerIdentity{"url": url}
}

// String returns the string representation of an identity
func (i ConsumerIdentity) String() string {
	data, _ := json.Marshal(i)
	return string(data)
}

// Key returns the object digest of an identity
func (i ConsumerIdentity) Key() []byte {
	data, _ := json.Marshal(i)
	return data
}

// Equals compares two identities
func (i ConsumerIdentity) Equals(o ConsumerIdentity) bool {
	if len(i) != len(o) {
		return false
	}

	for k, v := range i {
		if v2, ok := o[k]; !ok || v != v2 {
			return false
		}
	}
	return true
}

// Match implements the selector interface.
func (i ConsumerIdentity) Match(obj map[string]string) (bool, error) {
	for k, v := range i {
		if obj[k] != v {
			return false, nil
		}
	}
	return true, nil
}

// Copy copies identity
func (l ConsumerIdentity) Copy() ConsumerIdentity {
	if l == nil {
		return nil
	}
	n := ConsumerIdentity{}
	for k, v := range l {
		n[k] = v
	}
	return n
}
