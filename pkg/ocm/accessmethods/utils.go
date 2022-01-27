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

package accessmethods

import (
	"io"
	"io/ioutil"
	"sync"

	"github.com/gardener/ocm/pkg/common/accessio"
	"github.com/gardener/ocm/pkg/ocm/cpi"
	"github.com/gardener/ocm/pkg/runtime"
	"github.com/opencontainers/go-digest"
)

type AccessImplementation interface {
	Open() (io.ReadCloser, error)
	Size() int64
	MimeType() string
}

func GetImplementation(m cpi.AccessMethod) AccessImplementation {
	if d, ok := m.(*DefaultAccessMethod); ok {
		return d.impl
	}
	return nil
}

type DefaultAccessMethod struct {
	runtime.ObjectVersionedType
	lock   sync.Mutex
	impl   AccessImplementation
	digest digest.Digest
	size   int64
}

var _ cpi.AccessMethod = &DefaultAccessMethod{}

func NewDefaultAccessMethod(typ string, impl AccessImplementation) cpi.AccessMethod {
	return &DefaultAccessMethod{
		ObjectVersionedType: runtime.NewVersionedObjectType(typ),
		impl:                impl,
		size:                -1,
	}
}

func (m *DefaultAccessMethod) Digest() digest.Digest {
	m.lock.Lock()
	defer m.lock.Unlock()
	if m.digest == "" {
		reader, err := m.impl.Open()
		defer reader.Close()
		if err == nil {
			count := accessio.NewCountingReader(reader)
			m.digest, err = digest.Canonical.FromReader(count)
			if err == nil {
				m.size = count.Size()
			}
		}
	}
	return m.digest
}

func (m *DefaultAccessMethod) Size() int64 {
	m.lock.Lock()
	defer m.lock.Unlock()
	if m.size < 0 {
		m.size = m.impl.Size()
		if m.size < 0 {
			reader, err := m.impl.Open()
			if err == nil {
				defer reader.Close()
				var buf [8000]byte
				count := accessio.NewCountingReader(reader)
				for err == nil {
					_, err = count.Read(buf[:])
				}
				if err == io.EOF {
					m.size = count.Size()
				}
			}
		}
	}
	return m.size
}

func (m *DefaultAccessMethod) Get() ([]byte, error) {
	file, err := m.impl.Open()
	if err != nil {
		return nil, err
	}
	defer file.Close()
	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	m.lock.Lock()
	defer m.lock.Unlock()
	if m.digest == "" {
		m.digest = digest.FromBytes(data)
		m.size = int64(len(data))
	}
	return data, nil
}

func (m *DefaultAccessMethod) Reader() (io.ReadCloser, error) {
	return m.impl.Open()
}

func (m *DefaultAccessMethod) MimeType() string {
	return m.impl.MimeType()
}
