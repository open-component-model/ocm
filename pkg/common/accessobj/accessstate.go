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
	"sync"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/modern-go/reflect2"
	"github.com/opencontainers/go-digest"
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

var ErrReadOnly = accessio.ErrReadOnly

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
	Get() (accessio.BlobAccess, error)
	//Digest() digest.Digest
	Put(data []byte) error
}

// BlobStateAccess provides state handling for data given by a blob access
type BlobStateAccess struct {
	lock sync.RWMutex
	blob accessio.BlobAccess
}

var _ StateAccess = (*BlobStateAccess)(nil)

func NewBlobStateAccess(blob accessio.BlobAccess) StateAccess {
	return &BlobStateAccess{
		blob: blob,
	}
}

func NewBlobStateAccessForData(mimeType string, data []byte) StateAccess {
	return &BlobStateAccess{
		blob: accessio.BlobAccessForData(mimeType, data),
	}
}

func (b *BlobStateAccess) Get() (accessio.BlobAccess, error) {
	b.lock.RLock()
	defer b.lock.RUnlock()
	return b.blob, nil
}

func (b *BlobStateAccess) Put(data []byte) error {
	b.lock.Lock()
	defer b.lock.Unlock()
	b.blob = accessio.BlobAccessForData(b.blob.MimeType(), data)
	return nil
}

func (b *BlobStateAccess) Digest() digest.Digest {
	b.lock.RLock()
	defer b.lock.RUnlock()
	return b.blob.Digest()
}

// State manages the modification and access of state
// with a technical representation as byte array
// It tries to keep the byte representation unchanged as long as
// possible
type State interface {
	IsReadOnly() bool
	IsCreate() bool

	GetOriginalBlob() accessio.BlobAccess
	GetBlob() (accessio.BlobAccess, error)

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
	originalBlob accessio.BlobAccess
	original     interface{}
	current      interface{}
}

var _ State = (*state)(nil)

// NewState creates a new State based on its persistence handling
// and the management of its technical representation as byte array
func NewState(mode AccessMode, a StateAccess, p StateHandler) (State, error) {
	state, err := newState(mode, a, p)
	// avoid nil pinter problem: go is great
	if err != nil {
		return nil, err
	}
	return state, nil
}

func newState(mode AccessMode, a StateAccess, p StateHandler) (*state, error) {
	blob, err := a.Get()
	if err != nil {
		if (mode&ACC_CREATE) == 0 || !errors.IsErrNotFound(err) {
			return nil, err
		}
	}
	var current interface{}
	var original interface{}
	if blob == nil {
		current = p.Initial()
	} else {
		data, err := blob.Get()
		if err != nil {
			return nil, err
		}
		blob = accessio.BlobAccessForData(blob.MimeType(), data) // cache orginal data
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
		originalBlob: blob,
		original:     original,
		current:      current,
	}, nil
}

// NewBlobState provides state handling for and object persisted as a blob.
// It tries to keep the blob representation unchanged as long as possible
// consulting the state handler responsible for analysing the binary blob data
// and the object.
func NewBlobStateForBlob(mode AccessMode, blob accessio.BlobAccess, p StateHandler) (State, error) {
	if blob == nil {
		data, err := p.Encode(p.Initial())
		if err != nil {
			return nil, err
		}
		blob = accessio.BlobAccessForData("", data)
	}
	return NewState(mode, NewBlobStateAccess(blob), p)
}

// NewBlobStateForObject returns a representation state handling for a given object
func NewBlobStateForObject(mode AccessMode, obj interface{}, p StateHandler) (State, error) {
	if reflect2.IsNil(obj) {
		obj = p.Initial()
	}
	data, err := p.Encode(obj)
	if err != nil {
		return nil, err
	}
	return NewBlobStateForBlob(mode, accessio.BlobAccessForData("", data), p)
}

func (s *state) IsReadOnly() bool {
	return s.mode.IsReadonly()
}

func (s *state) IsCreate() bool {
	return s.mode.IsCreate()
}

func (s *state) Refresh() error {
	n, err := newState(s.mode, s.access, s.handler)
	if err != nil {
		return err
	}
	*s = *n
	return nil
}

func (s *state) GetOriginalState() interface{} {
	if s.originalBlob == nil {
		return nil
	}
	// always provide a private copy to not corrupt the internal state
	var original interface{}
	data, err := s.originalBlob.Get()
	if err == nil {
		original, err = s.handler.Decode(data)
	}
	if err != nil {
		panic("use of invalid state: " + err.Error())
	}
	return original
}

func (s *state) GetState() interface{} {
	return s.current
}

func (s *state) GetOriginalBlob() accessio.BlobAccess {
	return s.originalBlob
}

func (s *state) HasChanged() bool {
	return s.handler.Equivalent(s.original, s.current)
}

func (s *state) GetBlob() (accessio.BlobAccess, error) {
	if s.handler.Equivalent(s.original, s.current) {
		return s.originalBlob, nil
	}
	data, err := s.handler.Encode(s.current)
	if err != nil {
		return nil, err
	}
	if s.originalBlob != nil {
		return accessio.BlobAccessForData(s.originalBlob.MimeType(), data), nil
	}
	return accessio.BlobAccessForData("", data), nil
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
	mimeType := ""
	if s.originalBlob != nil {
		mimeType = s.originalBlob.MimeType()
	}
	s.originalBlob = accessio.BlobAccessForData(mimeType, data)
	s.original = original
	return true, nil
}

////////////////////////////////////////////////////////////////////////////////

type fileBasedAccess struct {
	filesystem vfs.FileSystem
	path       string
	mimeType   string
	mode       vfs.FileMode
}

func (f *fileBasedAccess) Get() (accessio.BlobAccess, error) {
	ok, err := vfs.FileExists(f.filesystem, f.path)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.ErrNotFoundWrap(vfs.ErrNotExist, "file", f.path)
	}
	return accessio.BlobAccessForFile(f.mimeType, f.path, f.filesystem), nil
}

func (f *fileBasedAccess) Put(data []byte) error {
	if err := vfs.WriteFile(f.filesystem, f.path, data, f.mode); err != nil {
		return fmt.Errorf("unable to write file %q: %w", f.path, err)
	}
	return nil
}

func (f *fileBasedAccess) Digest() digest.Digest {
	data, err := f.filesystem.Open(f.path)
	if err == nil {
		defer data.Close()
		d, err := digest.FromReader(data)
		if err == nil {
			return d
		}
	}
	return ""
}

////////////////////////////////////////////////////////////////////////////////

// NewFileBasedState create a new State object based on a file based persistence
// of the state carrying object.
func NewFileBasedState(acc AccessMode, fs vfs.FileSystem, path string, mimeType string, h StateHandler, mode vfs.FileMode) (State, error) {
	return NewState(acc, &fileBasedAccess{
		filesystem: fs,
		path:       path,
		mode:       mode,
		mimeType:   mimeType,
	}, h)
}
