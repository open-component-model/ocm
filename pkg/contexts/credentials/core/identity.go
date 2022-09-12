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

package core

import (
	"encoding/json"
	"sort"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
)

func init() {
	StandardIdentityMatchers.Register("partial", PartialMatch, "complete match of given pattern ignoring additional attributes")
	StandardIdentityMatchers.Register("exact", CompleteMatch, "exact match of given pattern set")
}

// IdentityMatcher checks whether id matches against pattern and if this match
// is better than the one for cur.
type IdentityMatcher func(pattern, cur, id ConsumerIdentity) bool

func CompleteMatch(pattern, cur, id ConsumerIdentity) bool {
	return pattern.Equals(id)
}

func NoMatch(pattern, cur, id ConsumerIdentity) bool {
	return false
}

func PartialMatch(pattern, cur, id ConsumerIdentity) bool {
	for k, v := range pattern {
		if c, ok := id[k]; !ok || c != v {
			return false
		}
	}
	return len(cur) == 0 || len(id) < len(cur)
}

func mergeMatcher(no IdentityMatcher, merge func([]IdentityMatcher) IdentityMatcher, matchers []IdentityMatcher) IdentityMatcher {
	var list []IdentityMatcher
	for _, m := range matchers {
		if m != nil {
			list = append(list, m)
		}
	}
	switch len(list) {
	case 0:
		return no
	case 1:
		return list[0]
	default:
		return merge(list)
	}
}

func defaultMatcher(matchers ...IdentityMatcher) IdentityMatcher {
	return mergeMatcher(nil, andMatcher, matchers)
}

func AndMatcher(matchers ...IdentityMatcher) IdentityMatcher {
	return mergeMatcher(NoMatch, andMatcher, matchers)
}

func OrMatcher(matchers ...IdentityMatcher) IdentityMatcher {
	return mergeMatcher(NoMatch, orMatcher, matchers)
}

func andMatcher(list []IdentityMatcher) IdentityMatcher {
	return func(pattern, cur, id ConsumerIdentity) bool {
		result := false
		for _, m := range list {
			if m != nil && !m(pattern, cur, id) {
				return false
			}
			result = true
		}
		return result
	}
}

func orMatcher(list []IdentityMatcher) IdentityMatcher {
	return func(pattern, cur, id ConsumerIdentity) bool {
		for _, m := range list {
			if m != nil && m(pattern, cur, id) {
				return true
			}
		}
		return false
	}
}

////////////////////////////////////////////////////////////////////////////////

// ConsumerIdentity describes the identity of a credential consumer.
type ConsumerIdentity map[string]string

// IdentityByURL return a simple url identity.
func IdentityByURL(url string) ConsumerIdentity {
	return ConsumerIdentity{"url": url}
}

// String returns the string representation of an identity.
func (i ConsumerIdentity) String() string {
	data, err := json.Marshal(i)
	if err != nil {
		logrus.Error(err)
	}
	return string(data)
}

// Key returns the object digest of an identity.
func (i ConsumerIdentity) Key() []byte {
	data, err := json.Marshal(i)
	if err != nil {
		logrus.Error(err)
	}
	return data
}

// Equals compares two identities.
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
func (i ConsumerIdentity) Match(obj map[string]string) bool {
	for k, v := range i {
		if obj[k] != v {
			return false
		}
	}
	return true
}

// Copy copies identity.
func (i ConsumerIdentity) Copy() ConsumerIdentity {
	if i == nil {
		return nil
	}
	n := ConsumerIdentity{}
	for k, v := range i {
		n[k] = v
	}
	return n
}

// SetNonEmptyValue sets a key-value pair only if the value is not empty.
func (i ConsumerIdentity) SetNonEmptyValue(name, value string) {
	if value != "" {
		i[name] = value
	}
}

////////////////////////////////////////////////////////////////////////////////

type IdentityMatcherInfo struct {
	Type        string
	Matcher     IdentityMatcher
	Description string
}

type IdentityMatcherInfos []IdentityMatcherInfo

func (l IdentityMatcherInfos) Size() int                { return len(l) }
func (l IdentityMatcherInfos) Key(i int) string         { return l[i].Type }
func (l IdentityMatcherInfos) Description(i int) string { return l[i].Description }

type IdentityMatcherRegistry interface {
	Register(typ string, matcher IdentityMatcher, desc string)
	Get(typ string) (IdentityMatcher, string)
	List() IdentityMatcherInfos
}

type defaultMatchers struct {
	lock  sync.Mutex
	types map[string]IdentityMatcherInfo
}

func NewMatcherRegistry() IdentityMatcherRegistry {
	return &defaultMatchers{types: map[string]IdentityMatcherInfo{}}
}

func (r *defaultMatchers) Register(typ string, matcher IdentityMatcher, desc string) {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.types[typ] = IdentityMatcherInfo{typ, matcher, desc}
}

func (r *defaultMatchers) Get(typ string) (IdentityMatcher, string) {
	r.lock.Lock()
	defer r.lock.Unlock()
	i, ok := r.types[typ]
	if !ok {
		return nil, ""
	}
	return i.Matcher, i.Description
}

func (r *defaultMatchers) List() IdentityMatcherInfos {
	r.lock.Lock()
	defer r.lock.Unlock()
	var list IdentityMatcherInfos

	for _, i := range r.types {
		list = append(list, i)
	}

	sort.Slice(list, func(i, j int) bool { return strings.Compare(list[i].Type, list[j].Type) < 0 })
	return list
}

var StandardIdentityMatchers = NewMatcherRegistry()

func RegisterIdentityMatcher(typ string, matcher IdentityMatcher, desc string) {
	StandardIdentityMatchers.Register(typ, matcher, desc)
}
