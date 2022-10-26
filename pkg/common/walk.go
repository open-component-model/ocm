// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"github.com/open-component-model/ocm/pkg/utils"
)

type NameVersionInfo map[NameVersion]interface{}

func (s NameVersionInfo) Add(nv NameVersion, data ...interface{}) bool {
	if _, ok := s[nv]; !ok {
		s[nv] = utils.Optional(data...)
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
