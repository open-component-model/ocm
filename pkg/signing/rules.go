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
	Field(name string, value interface{}) (string, interface{}, ExcludeRules)
	Element(v interface{}) (bool, interface{}, ExcludeRules)
}

////////////////////////////////////////////////////////////////////////////////

type NoExcludes struct{}

var _ ExcludeRules = NoExcludes{}

func (r NoExcludes) Field(name string, value interface{}) (string, interface{}, ExcludeRules) {
	return name, value, r
}

func (r NoExcludes) Element(value interface{}) (bool, interface{}, ExcludeRules) {
	return false, value, r
}

////////////////////////////////////////////////////////////////////////////////

type MapIncludes map[string]ExcludeRules

var _ ExcludeRules = MapIncludes{}

func (r MapIncludes) Field(name string, value interface{}) (string, interface{}, ExcludeRules) {
	c, ok := r[name]
	if ok {
		if c == nil {
			c = NoExcludes{}
		}
		return name, value, c
	}
	return "", nil, nil
}

func (r MapIncludes) Element(v interface{}) (bool, interface{}, ExcludeRules) {
	panic("invalid exclude structure, require arry but found struct rules")
}

////////////////////////////////////////////////////////////////////////////////

type MapExcludes map[string]ExcludeRules

var _ ExcludeRules = MapExcludes{}

func (r MapExcludes) Field(name string, value interface{}) (string, interface{}, ExcludeRules) {
	c, ok := r[name]
	if ok {
		if c == nil {
			return "", nil, nil
		}
	} else {
		c = NoExcludes{}
	}
	return name, value, c
}

func (r MapExcludes) Element(value interface{}) (bool, interface{}, ExcludeRules) {
	panic("invalid exclude structure, require arry but found struct rules")
}

////////////////////////////////////////////////////////////////////////////////

type ValueMapper func(v interface{}) interface{}
type ValueChecker func(value interface{}) bool

type DynamicInclude struct {
	ValueChecker ValueChecker
	ValueMapper  ValueMapper
	Continue     ExcludeRules
	Name         string
}

func (r *DynamicInclude) Check(value interface{}) bool {
	return r == nil || r.ValueChecker == nil || r.ValueChecker(value)
}

type DynamicMapIncludes map[string]*DynamicInclude

var _ ExcludeRules = DynamicMapIncludes{}

func (r DynamicMapIncludes) Field(name string, value interface{}) (string, interface{}, ExcludeRules) {
	e, ok := r[name]
	if ok && e.Check(value) {
		var c ExcludeRules = NoExcludes{}
		if e != nil {
			if e.Name != "" {
				name = e.Name
			}
			if e.Continue != nil {
				c = e.Continue
			}
			if e.ValueMapper != nil {
				value = e.ValueMapper(value)
			}
		}
		return name, value, c
	}
	return "", nil, nil
}

func (r DynamicMapIncludes) Element(value interface{}) (bool, interface{}, ExcludeRules) {
	panic("invalid exclude structure, require arry but found struct rules")
}

////////////////////////////////////////////////////////////////////////////////

type DynamicExclude struct {
	ValueChecker ValueChecker
	ValueMapper  ValueMapper
	Continue     ExcludeRules
	Name         string
}

func (r *DynamicExclude) Check(value interface{}) bool {
	return r == nil || (r.ValueChecker != nil && r.ValueChecker(value)) || (r.ValueChecker == nil && r.Continue == nil)
}

type DynamicMapExcludes map[string]*DynamicExclude

var _ ExcludeRules = DynamicMapExcludes{}

func (r DynamicMapExcludes) Field(name string, value interface{}) (string, interface{}, ExcludeRules) {
	var c ExcludeRules
	e, ok := r[name]
	if ok {
		if e.Check(value) {
			return "", nil, nil
		}
		if e.Name != "" {
			name = e.Name
		}
		c = e.Continue
	} else {
		c = NoExcludes{}
	}
	if e != nil && e.ValueMapper != nil {
		value = e.ValueMapper(value)
	}
	return name, value, c
}

func (r DynamicMapExcludes) Element(value interface{}) (bool, interface{}, ExcludeRules) {
	panic("invalid exclude structure, require arry but found struct rules")
}

////////////////////////////////////////////////////////////////////////////////

type DynamicArrayExcludes struct {
	ValueChecker ValueChecker
	ValueMapper  ValueMapper
	Continue     ExcludeRules
}

var _ ExcludeRules = DynamicArrayExcludes{}

func (r DynamicArrayExcludes) Field(name string, value interface{}) (string, interface{}, ExcludeRules) {
	panic("invalid exclude structure, require struct but found array rules")
}

func (r DynamicArrayExcludes) Element(value interface{}) (bool, interface{}, ExcludeRules) {
	excl := r.Check(value)
	if !excl && r.ValueMapper != nil {
		value = r.ValueMapper(value)
	}
	if excl || r.Continue != nil {
		return excl, value, r.Continue
	}
	return false, value, NoExcludes{}
}

func (r DynamicArrayExcludes) Check(value interface{}) bool {
	return r.Continue == nil || (r.ValueChecker != nil && r.ValueChecker(value))
}

////////////////////////////////////////////////////////////////////////////////

type ArrayExcludes struct {
	Continue ExcludeRules
}

var _ ExcludeRules = ArrayExcludes{}

func (r ArrayExcludes) Field(name string, value interface{}) (string, interface{}, ExcludeRules) {
	panic("invalid exclude structure, require struct but found array rules")
}

func (r ArrayExcludes) Element(value interface{}) (bool, interface{}, ExcludeRules) {
	return false, value, r.Continue
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
