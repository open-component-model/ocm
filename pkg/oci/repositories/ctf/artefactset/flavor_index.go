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

package artefactset

import (
	"sync"

	"github.com/gardener/ocm/pkg/common/accessio"
	"github.com/gardener/ocm/pkg/errors"
	"github.com/gardener/ocm/pkg/oci/artdesc"
	"github.com/gardener/ocm/pkg/oci/core"
	"github.com/gardener/ocm/pkg/oci/cpi"
	"github.com/opencontainers/go-digest"
)

type Index struct {
	access       ArtefactSetContainer
	index        *artdesc.Index
	lock         sync.RWMutex
	*BlobHandler // Index offers all: blobs, artefacts, manifests and indices
}

var _ cpi.IndexAccess = (*Index)(nil)

func NewIndex(access ArtefactSetContainer, def ...*artdesc.Index) core.IndexAccess {
	var index *artdesc.Index
	if len(def) == 0 || def[0] == nil {
		index = artdesc.NewIndex()
	} else {
		index = def[0]
	}
	i := &Index{
		access: access,
		index:  index,
	}
	i.BlobHandler = NewBlobHandler(access, i)
	return i
}

func (i *Index) IsManifest() bool {
	return false
}

func (i *Index) IsIndex() bool {
	return true
}

func (i *Index) Manifest() (*artdesc.Manifest, error) {
	return nil, errors.ErrInvalid()
}

func (i *Index) Index() (*artdesc.Index, error) {
	return i.index, nil
}

func (i *Index) Artefact() *artdesc.Artefact {
	a := artdesc.New()
	_ = a.SetIndex(i.index)
	return a
}

func (i *Index) GetDescriptor() *artdesc.Index {
	return i.index
}

func (i *Index) GetBlobDescriptor(digest digest.Digest) *cpi.Descriptor {
	d := i.index.GetBlobDescriptor(digest)
	if d != nil {
		return d
	}
	return i.access.GetBlobDescriptor(digest)
}

func (a *Index) AddArtefact(art cpi.Artefact, platform *artdesc.Platform) (access accessio.BlobAccess, err error) {
	blob, err := a.access.AddArtefact(art, platform)
	if err != nil {
		return nil, err
	}

	a.lock.Lock()
	defer a.lock.Unlock()

	a.index.Manifests = append(a.index.Manifests, cpi.Descriptor{
		MediaType:   blob.MimeType(),
		Digest:      blob.Digest(),
		Size:        blob.Size(),
		URLs:        nil,
		Annotations: nil,
		Platform:    platform,
	})
	return blob, nil
}
