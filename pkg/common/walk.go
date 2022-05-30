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

type NameVersionInfo map[NameVersion]interface{}

func (s NameVersionInfo) Add(nv NameVersion, data ...interface{}) bool {
	var d interface{}
	if len(data) > 0 {
		d = data[0]
	}
	if _, ok := s[nv]; !ok {
		s[nv] = d
		return true
	}
	return false
}

func (s NameVersionInfo) Contains(nv NameVersion) bool {
	_, ok := s[nv]
	return ok
}

type WalkingState struct {
	Closure NameVersionInfo
	History History
}

func NewWalkingState() WalkingState {
	return WalkingState{Closure: NameVersionInfo{}}
}

func (s *WalkingState) Add(kind string, nv NameVersion) (bool, error) {
	if err := s.History.Add(kind, nv); err != nil {
		return false, err
	}
	return s.Closure.Add(nv), nil
}

func (s *WalkingState) Contains(nv NameVersion) bool {
	_, ok := s.Closure[nv]
	return ok
}
