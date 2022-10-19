// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package docker

import (
	"sync"

	"github.com/containers/image/v5/types"
	"github.com/opencontainers/go-digest"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/contexts/oci/cpi"
	"github.com/open-component-model/ocm/pkg/errors"
)

type dockerSource struct {
	lock     sync.RWMutex
	src      types.ImageSource
	img      types.Image
	refcount int
}

var _ accessio.BlobSource = (*dockerSource)(nil)

func newDockerSource(img types.Image, src types.ImageSource) *dockerSource {
	return &dockerSource{
		src:      src,
		img:      img,
		refcount: 1,
	}
}

func (c *dockerSource) Ref() error {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.refcount == 0 {
		return accessio.ErrClosed
	}
	c.refcount++
	return nil
}

func (c *dockerSource) Unref() error {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.refcount == 0 {
		return accessio.ErrClosed
	}
	c.refcount--
	return c.src.Close()
}

func (d *dockerSource) GetBlobData(digest digest.Digest) (int64, accessio.DataAccess, error) {
	info := d.img.ConfigInfo()
	if info.Digest == digest {
		data, err := d.img.ConfigBlob(dummyContext)
		if err != nil {
			return -1, nil, err
		}
		return info.Size, accessio.DataAccessForBytes(data), nil
	}
	info.Digest = ""
	for _, l := range d.img.LayerInfos() {
		if l.Digest == digest {
			info = l
			acc, err := NewDataAccess(d.src, info, false)
			return l.Size, acc, err
		}
	}
	return -1, nil, cpi.ErrBlobNotFound(digest)
}

////////////////////////////////////////////////////////////////////////////////

type daemonArtefactProvider struct {
	lock      sync.Mutex
	namespace *NamespaceContainer
	cache     accessio.BlobCache
}

var _ cpi.ArtefactProvider = (*daemonArtefactProvider)(nil)

func (d *daemonArtefactProvider) IsClosed() bool {
	d.lock.Lock()
	defer d.lock.Unlock()
	return d.cache == nil
}

func (d *daemonArtefactProvider) IsReadOnly() bool {
	return d.namespace.IsReadOnly()
}

func (d *daemonArtefactProvider) GetBlobDescriptor(digest digest.Digest) *cpi.Descriptor {
	return nil
}

func (d *daemonArtefactProvider) Close() error {
	d.lock.Lock()
	defer d.lock.Unlock()

	if d.cache != nil {
		err := d.cache.Unref()
		d.cache = nil
		return err
	}
	return nil
}

func (d *daemonArtefactProvider) GetBlobData(digest digest.Digest) (int64, cpi.DataAccess, error) {
	return d.cache.GetBlobData(digest)
}

func (d *daemonArtefactProvider) GetArtefact(digest digest.Digest) (cpi.ArtefactAccess, error) {
	return nil, errors.ErrInvalid()
}

func (d *daemonArtefactProvider) AddBlob(access cpi.BlobAccess) error {
	_, _, err := d.cache.AddBlob(access)
	return err
}

func (d *daemonArtefactProvider) AddArtefact(art cpi.Artefact) (access accessio.BlobAccess, err error) {
	return nil, errors.ErrInvalid()
}
