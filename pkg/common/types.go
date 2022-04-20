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

package common

import (
	"fmt"

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

func (n NameVersion) String() string {
	if n.version == "" {
		return n.name
	}
	if n.name == "" {
		return n.version
	}
	return n.name + ":" + n.version
}

////////////////////////////////////////////////////////////////////////////////

type HistorySource interface {
	GetHistory() History
}

type History []NameVersion

func (h History) String() string {
	s := ""
	sep := ""
	for _, e := range h {
		s = fmt.Sprintf("%s%s%s", s, sep, e)
		sep = "->"
	}
	return s
}

func (h History) Contains(nv NameVersion) bool {
	for _, e := range h {
		if e == nv {
			return true
		}
	}
	return false
}

func (h History) HasPrefix(o History) bool {
	if len(o) > len(h) {
		return false
	}
	for i, e := range o {
		if e != h[i] {
			return false
		}
	}
	return true
}

func (h History) Equals(o History) bool {
	if len(h) != len(o) {
		return false
	}
	if h == nil || o == nil {
		return false
	}

	for i, e := range h {
		if e != o[i] {
			return false
		}
	}
	return true
}

func (h *History) Add(kind string, nv NameVersion) error {
	if h.Contains(nv) {
		return errors.ErrRecusion(kind, nv, *h)
	}
	*h = append(*h, nv)
	return nil
}

func (h History) Copy() History {
	return append(h[:0:0], h...)
}

func (h History) RemovePrefix(prefix History) History {
	for i, e := range prefix {
		if len(h) <= i || e != h[i] {
			return h[i:]
		}
	}
	return h[len(prefix):]
}
