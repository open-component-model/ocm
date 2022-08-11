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
	"fmt"
	"strings"

	"github.com/open-component-model/ocm/pkg/errors"
)

// VersionedElement describes an element that has a name and a version
type VersionedElement interface {
	// GetName gets the name of the element
	GetName() string
	// GetVersion gets the version of the element
	GetVersion() string
}

type NameVersion struct {
	name    string
	version string
}

var _ VersionedElement = (*NameVersion)(nil)

func NewNameVersion(name, version string) NameVersion {
	return NameVersion{name, version}
}

func VersionedElementKey(v VersionedElement) NameVersion {
	if k, ok := v.(NameVersion); ok {
		return k
	}
	return NameVersion{v.GetName(), v.GetVersion()}
}

func (n NameVersion) GetName() string {
	return n.name
}

func (n NameVersion) GetVersion() string {
	return n.version
}

func (n NameVersion) MarshalJSON() ([]byte, error) {
	return json.Marshal(fmt.Sprintf("%s:%s", n.GetName(), n.GetVersion()))
}

func (n NameVersion) Compare(o NameVersion) int {
	c := strings.Compare(n.name, o.name)
	if c == 0 {
		return strings.Compare(n.version, o.version)
	}
	return c
}

func (n NameVersion) String() string {
	if n.version == "" {
		return n.name
	}
	if n.name == "" {
		return n.version
	}
	return n.name + ":" + n.version
}

func ParseNameVersion(s string) (NameVersion, error) {
	a := strings.Split(s, ":")
	if len(a) != 2 {
		return NameVersion{}, errors.ErrInvalid("name:version", s)
	}
	return NewNameVersion(strings.TrimSpace(a[0]), strings.TrimSpace(a[1])), nil
}
