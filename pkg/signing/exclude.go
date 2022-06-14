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

package signing

// ExcludeRules defines the rules for normalization excludes
type ExcludeRules interface {
	Field(name string, value interface{}) (string, ExcludeRules)
	Element(v interface{}) (bool, ExcludeRules)
}

////////////////////////////////////////////////////////////////////////////////

type NoExcludes struct{}

var _ ExcludeRules = NoExcludes{}

func (r NoExcludes) Field(name string, value interface{}) (string, ExcludeRules) {
	return name, r
}

func (r NoExcludes) Element(v interface{}) (bool, ExcludeRules) {
	return false, r
}

////////////////////////////////////////////////////////////////////////////////

type MapExcludes map[string]ExcludeRules

var _ ExcludeRules = MapExcludes{}

func (r MapExcludes) Field(name string, value interface{}) (string, ExcludeRules) {
	e, ok := r[name]
	if ok {
		if e == nil {
			return "", nil
		}
	} else {
		e = NoExcludes{}
	}
	return name, e
}

func (r MapExcludes) Element(v interface{}) (bool, ExcludeRules) {
	panic("invalid exclude structure, require arry but found struct rules")
}

////////////////////////////////////////////////////////////////////////////////

type DynamicExclude struct {
	ValueChecker ValueChecker
	Continue     ExcludeRules
	Name         string
}

func (r DynamicExclude) Check(value interface{}) bool {
	return r.Continue == nil || (r.ValueChecker != nil && r.ValueChecker(value))
}

type DynamicMapExcludes map[string]DynamicExclude

type ValueChecker func(value interface{}) bool

var _ ExcludeRules = DynamicMapExcludes{}

func (r DynamicMapExcludes) Field(name string, value interface{}) (string, ExcludeRules) {
	e, ok := r[name]
	if ok {
		if e.Check(value) {
			return "", nil
		}
		if e.Name != "" {
			name = e.Name
		}
	}
	return name, NoExcludes{}
}

func (r DynamicMapExcludes) Element(v interface{}) (bool, ExcludeRules) {
	panic("invalid exclude structure, require arry but found struct rules")
}

////////////////////////////////////////////////////////////////////////////////

type DynamicArrayExcludes struct {
	ValueChecker ValueChecker
	Continue     ExcludeRules
}

var _ ExcludeRules = DynamicArrayExcludes{}

func (r DynamicArrayExcludes) Field(name string, value interface{}) (string, ExcludeRules) {
	panic("invalid exclude structure, require struct but found array rules")
}

func (r DynamicArrayExcludes) Element(v interface{}) (bool, ExcludeRules) {
	excl := r.Check(v)
	if excl || r.Continue != nil {
		return excl, r.Continue
	}
	return false, NoExcludes{}
}

func (r DynamicArrayExcludes) Check(value interface{}) bool {
	return r.Continue == nil || (r.ValueChecker != nil && r.ValueChecker(value))
}

////////////////////////////////////////////////////////////////////////////////

type ArrayExcludes struct {
	Elements ExcludeRules
}

var _ ExcludeRules = ArrayExcludes{}

func (r ArrayExcludes) Field(name string, value interface{}) (string, ExcludeRules) {
	panic("invalid exclude structure, require struct but found array rules")
}

func (r ArrayExcludes) Element(v interface{}) (bool, ExcludeRules) {
	return false, r.Elements
}

////////////////////////////////////////////////////////////////////////////////

func IgnoreResourcesWithNoneAccess(v interface{}) bool {
	return CheckIgnoreResourcesWithAccessType("none", v)
}

func IgnoreResourcesWithAccessType(t string) func(v interface{}) bool {
	return func(v interface{}) bool {
		return CheckIgnoreResourcesWithAccessType(t, v)
	}
}

func CheckIgnoreResourcesWithAccessType(t string, v interface{}) bool {
	access := v.(map[string]interface{})["access"]
	if access == nil {
		return true
	}
	typ := access.(map[string]interface{})["type"]
	return typ == t
}
