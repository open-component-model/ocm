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

package accessobj

import (
	"fmt"

	"github.com/gardener/ocm/pkg/errors"
	"github.com/mandelsoft/vfs/pkg/vfs"
)

// These objects deal with descriptor based state descriptions
// of an access object

type AccessMode byte

const (
	ACC_WRITABLE = AccessMode(0)
	ACC_READONLY = AccessMode(1)
	ACC_CREATE   = AccessMode(2)
)

func (m AccessMode) IsReadonly() bool {
	return (m & ACC_READONLY) != 0
}

func (m AccessMode) IsCreate() bool {
	return (m & ACC_CREATE) != 0
}

var ErrReadOnly = errors.ErrReadOnly()

// StateHandler is responsible to handle the technical representation of state
// carrying object as byte array.
type StateHandler interface {
	Initial() interface{}
	Encode(d interface{}) ([]byte, error)
	Decode([]byte) (interface{}, error)
	Equivalent(a, b interface{}) bool
}

// StateAccess is responsible to handle the persistence
// of a state object
type StateAccess interface {
	// Get returns the technical representation of a state object from its persistence
	// It MUST return an errors.IsErrNotFound compatible error
	// if the persistence not yet exists.
	Get() ([]byte, error)
	Put(data []byte) error
}

// State manages the modification and access of state
// with a technical representation as byte array
type State interface {
	IsReadOnly() bool
	IsCreate() bool

	GetOriginalData() []byte
	GetData() ([]byte, error)

	HasChanged() bool
	GetOriginalState() interface{}
	GetState() interface{}

	// Update updates the technical representation in its persistence
	Update() (bool, error)
}

type state struct {
	mode         AccessMode
	access       StateAccess
	handler      StateHandler
	originalData []byte
	original     interface{}
	current      interface{}
}

// NewState creates a new State based on its persistence handling
// and the management of its technical representation as byte array
func NewState(mode AccessMode, a StateAccess, p StateHandler) (*state, error) {
	data, err := a.Get()
	if err != nil {
		if (mode&ACC_CREATE) == 0 || !errors.IsErrNotFound(err) {
			return nil, err
		}
	}
	var current interface{}
	var original interface{}
	if data == nil {
		current = p.Initial()
	} else {
		current, err = p.Decode(data)
		if err != nil {
			return nil, err
		}
		// we don't need a copy operation, because we can just deserialize it twice.
		original, _ = p.Decode(data)
	}

	return &state{
		mode:         mode,
		access:       a,
		handler:      p,
		originalData: data,
		original:     original,
		current:      current,
	}, nil
}

func (s *state) IsReadOnly() bool {
	return s.mode.IsReadonly()
}

func (s *state) IsCreate() bool {
	return s.mode.IsCreate()
}

func (s *state) Refresh() error {
	n, err := NewState(s.mode, s.access, s.handler)
	if err != nil {
		return err
	}
	*s = *n
	return nil
}

func (s *state) GetOriginalState() interface{} {
	if s.originalData == nil {
		return nil
	}
	// always provide a private copy to not corrupt the internal state
	original, err := s.handler.Decode(s.originalData)
	if err != nil {
		panic("use of invalid state")
	}
	return original
}

func (s *state) GetState() interface{} {
	return s.current
}

func (s *state) GetOriginalData() []byte {
	return s.originalData
}

func (s *state) HasChanged() bool {
	return s.handler.Equivalent(s.original, s.current)
}

func (s *state) GetData() ([]byte, error) {
	if s.handler.Equivalent(s.original, s.current) {
		return s.originalData, nil
	}
	return s.handler.Encode(s.current)
}

func (s *state) Update() (bool, error) {
	if s.handler.Equivalent(s.original, s.current) {
		return false, nil
	}

	if s.IsReadOnly() {
		return true, ErrReadOnly
	}
	data, err := s.handler.Encode(s.current)
	if err != nil {
		return false, err
	}
	original, err := s.handler.Decode(data)
	if err != nil {
		return false, err
	}
	err = s.access.Put(data)
	if err != nil {
		return false, err
	}
	s.originalData = data
	s.original = original
	return true, nil
}

////////////////////////////////////////////////////////////////////////////////

type fileBasedAccess struct {
	filesystem vfs.FileSystem
	path       string
	mode       vfs.FileMode
}

func (f *fileBasedAccess) Get() ([]byte, error) {
	data, err := vfs.ReadFile(f.filesystem, f.path)
	if err != nil {
		if vfs.IsErrNotExist(err) {
			return nil, errors.ErrNotFoundWrap(err, "file", f.path)
		}
	}
	return data, err
}

func (f *fileBasedAccess) Put(data []byte) error {
	if err := vfs.WriteFile(f.filesystem, f.path, data, f.mode); err != nil {
		return fmt.Errorf("unable to write file %q: %w", f.path, err)
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////

// NewFileBasedState create a new State object based on a file based persistence
// of the state carrying object.
func NewFileBasedState(acc AccessMode, fs vfs.FileSystem, path string, h StateHandler, mode vfs.FileMode) (State, error) {
	return NewState(acc, &fileBasedAccess{fs, path, mode}, h)

}
