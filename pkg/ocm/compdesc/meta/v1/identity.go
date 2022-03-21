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

package v1

import (
	"encoding/json"
	"fmt"

	"k8s.io/apimachinery/pkg/util/validation/field"
)

// Identity describes the identity of an object.
// Only ascii characters are allowed
// +k8s:deepcopy-gen=true
// +k8s:openapi-gen=true
type Identity map[string]string

// NewIdentity return a simple name identity
func NewIdentity(name string, extras ...string) Identity {
	id := Identity{"name": name}
	i := 0
	for i < len(extras) {
		if i+1 < len(extras) {
			id[extras[i]] = extras[i+1]
		} else {
			id[extras[i]] = ""
		}
		i += 2
	}
	return id
}

// Digest returns the object digest of an identity
func (i Identity) Digest() []byte {
	data, _ := json.Marshal(i)
	return data
}

// Equals compares two identities
func (i Identity) Equals(o Identity) bool {
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

func (l Identity) String() string {
	if l == nil {
		return ""
	}

	s := ""
	sep := ""
	for k, v := range l {
		s = fmt.Sprintf("%s%s%q=%q", s, sep, k, v)
		sep = ","
	}
	return s
}

// Match implements the selector interface.
func (i Identity) Match(obj map[string]string) (bool, error) {
	for k, v := range i {
		if obj[k] != v {
			return false, nil
		}
	}
	return true, nil
}

// Copy copies identity
func (l Identity) Copy() Identity {
	if l == nil {
		return nil
	}
	n := Identity{}
	for k, v := range l {
		n[k] = v
	}
	return n
}

// ValidateIdentity validates the identity of object.
func ValidateIdentity(fldPath *field.Path, id Identity) field.ErrorList {
	allErrs := field.ErrorList{}

	for key := range id {
		if key == SystemIdentityName {
			allErrs = append(allErrs, field.Forbidden(fldPath.Key(SystemIdentityName), "name is a reserved system identity label"))
		}

		if !IsASCII(key) {
			allErrs = append(allErrs, field.Forbidden(fldPath.Key(key), "key contains non-ascii characters"))
		}
		if !IsIdentity(key) {
			allErrs = append(allErrs, field.Invalid(fldPath.Key(key), key, IdentityKeyValidationErrMsg))
		}
	}
	return allErrs
}
