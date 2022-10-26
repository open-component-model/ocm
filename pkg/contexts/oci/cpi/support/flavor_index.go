// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package support

import (
	"github.com/opencontainers/go-digest"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/oci/artdesc"
	"github.com/open-component-model/ocm/pkg/contexts/oci/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/oci/internal"
	"github.com/open-component-model/ocm/pkg/errors"
)

type IndexImpl struct {
	artefactBase
}

var _ cpi.IndexAccess = (*IndexImpl)(nil)

type indexMapper struct {
	accessobj.State
}

var _ accessobj.State = (*indexMapper)(nil)

func (m *indexMapper) GetState() interface{} {
	return m.State.GetState().(*artdesc.Artefact).Index()
}

func (m *indexMapper) GetOriginalState() interface{} {
	return m.State.GetOriginalState().(*artdesc.Artefact).Index()
}

func NewIndexForArtefact(a *ArtefactImpl) *IndexImpl {
	m := &IndexImpl{
		artefactBase: artefactBase{
			container: a.container,
			state:     &indexMapper{a.state},
		},
	}
	return m
}

func (a *IndexImpl) NewArtefact(art ...*artdesc.Artefact) (cpi.ArtefactAccess, error) {
	return a.newArtefact(art...)
}

func (i *IndexImpl) AddBlob(blob internal.BlobAccess) error {
	return i.container.AddBlob(blob)
}

func (i *IndexImpl) Manifest() (*artdesc.Manifest, error) {
	return nil, errors.ErrInvalid()
}

func (i *IndexImpl) Index() (*artdesc.Index, error) {
	return i.GetDescriptor(), nil
}

func (i *IndexImpl) Artefact() *artdesc.Artefact {
	a := artdesc.New()
	_ = a.SetIndex(i.GetDescriptor())
	return a
}

func (i *IndexImpl) GetDescriptor() *artdesc.Index {
	return i.state.GetState().(*artdesc.Index)
}

func (i *IndexImpl) GetBlobDescriptor(digest digest.Digest) *cpi.Descriptor {
	d := i.GetDescriptor().GetBlobDescriptor(digest)
	if d != nil {
		return d
	}
	return i.container.GetBlobDescriptor(digest)
}

func (i *IndexImpl) GetBlob(digest digest.Digest) (internal.BlobAccess, error) {
	d := i.GetBlobDescriptor(digest)
	if d != nil {
		size, data, err := i.container.GetBlobData(digest)
		if err != nil {
			return nil, err
		}
		err = AdjustSize(d, size)
		if err != nil {
			return nil, err
		}
		return accessio.BlobAccessForDataAccess(d.Digest, d.Size, d.MediaType, data), nil
	}
	return nil, cpi.ErrBlobNotFound(digest)
}

func (i *IndexImpl) GetArtefact(digest digest.Digest) (internal.ArtefactAccess, error) {
	for _, d := range i.GetDescriptor().Manifests {
		if d.Digest == digest {
			return i.container.GetArtefact("@" + digest.String())
		}
	}
	return nil, errors.ErrNotFound(cpi.KIND_OCIARTEFACT, digest.String())
}

func (a *IndexImpl) AddArtefact(art cpi.Artefact, platform *artdesc.Platform) (access accessio.BlobAccess, err error) {
	blob, err := a.container.AddArtefact(art)
	if err != nil {
		return nil, err
	}

	a.lock.Lock()
	defer a.lock.Unlock()

	d := a.GetDescriptor()
	d.Manifests = append(d.Manifests, cpi.Descriptor{
		MediaType:   blob.MimeType(),
		Digest:      blob.Digest(),
		Size:        blob.Size(),
		URLs:        nil,
		Annotations: nil,
		Platform:    platform,
	})
	return blob, nil
}
