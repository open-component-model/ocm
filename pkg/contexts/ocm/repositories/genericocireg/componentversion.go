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

package genericocireg

import (
	"path"
	"strings"

	"github.com/opencontainers/go-digest"

	"github.com/open-component-model/ocm/pkg/common/accessio"

	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/oci"
	"github.com/open-component-model/ocm/pkg/contexts/oci/artdesc"
	"github.com/open-component-model/ocm/pkg/contexts/oci/repositories/artefactset"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/localblob"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/localociblob"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/ociartefact"
	ocihdlr "github.com/open-component-model/ocm/pkg/contexts/ocm/blobhandler/oci"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi/support"
	"github.com/open-component-model/ocm/pkg/errors"
)

type ComponentVersion struct {
	container *ComponentVersionContainer
	*support.ComponentVersionAccess
}

var _ cpi.ComponentVersionAccess = (*ComponentVersion)(nil)

// newComponentVersionAccess creates an component access for the artefact access, if this fails the artefact acess is closed.
func newComponentVersionAccess(mode accessobj.AccessMode, comp *componentAccessImpl, version string, access oci.ArtefactAccess) (*ComponentVersion, error) {
	c, err := newComponentVersionContainer(mode, comp, version, access)
	if err != nil {
		return nil, err
	}
	acc, err := support.NewComponentVersionAccess(c, true)
	if err != nil {
		c.Close()
		return nil, err
	}
	return &ComponentVersion{
		container:              c,
		ComponentVersionAccess: acc,
	}, nil
}

////////////////////////////////////////////////////////////////////////////////

type ComponentVersionContainer struct {
	comp     *ComponentAccess
	version  string
	access   oci.ArtefactAccess
	manifest oci.ManifestAccess
	state    accessobj.State
}

var _ support.ComponentVersionContainer = (*ComponentVersionContainer)(nil)

func newComponentVersionContainer(mode accessobj.AccessMode, comp *componentAccessImpl, version string, access oci.ArtefactAccess) (*ComponentVersionContainer, error) {
	m := access.ManifestAccess()
	if m == nil {
		return nil, errors.ErrInvalid("artefact type")
	}
	state, err := NewState(mode, comp.name, version, m)
	if err != nil {
		access.Close()
		return nil, err
	}
	v, err := comp.View(false)
	if err != nil {
		access.Close()
		return nil, err
	}
	return &ComponentVersionContainer{
		comp:     v,
		version:  version,
		access:   access,
		manifest: m,
		state:    state,
	}, nil
}

func (c *ComponentVersionContainer) Close() error {
	if c.manifest == nil {
		return accessio.ErrClosed
	}
	c.manifest = nil
	err := c.access.Close()
	if err != nil {
		c.comp.Close()
		return err
	}
	return c.comp.Close()
}

func (c *ComponentVersionContainer) Check() error {
	if c.version != c.GetDescriptor().Version {
		return errors.ErrInvalid("component version", c.GetDescriptor().Version)
	}
	if c.comp.name != c.GetDescriptor().Name {
		return errors.ErrInvalid("component name", c.GetDescriptor().Name)
	}
	return nil
}

func (c *ComponentVersionContainer) GetContext() cpi.Context {
	return c.comp.GetContext()
}

func (c *ComponentVersionContainer) ComponentAccess() cpi.ComponentAccess {
	return c.comp
}

func (c *ComponentVersionContainer) IsReadOnly() bool {
	return c.state.IsReadOnly()
}

func (c *ComponentVersionContainer) IsClosed() bool {
	return c.manifest == nil
}

func (c *ComponentVersionContainer) AccessMethod(a cpi.AccessSpec) (cpi.AccessMethod, error) {
	if a.GetKind() == localblob.Type {
		a, err := c.comp.GetContext().AccessSpecForSpec(a)
		if err != nil {
			return nil, err
		}
		return newLocalBlobAccessMethod(a.(*localblob.AccessSpec), c.comp.namespace, c.comp.GetContext())
	}
	if a.GetKind() == localociblob.Type {
		a, err := c.comp.GetContext().AccessSpecForSpec(a)
		if err != nil {
			return nil, err
		}
		return newLocalOCIBlobAccessMethod(a.(*localociblob.AccessSpec), c.comp.namespace)
	}
	return nil, errors.ErrNotSupported(errors.KIND_ACCESSMETHOD, a.GetType(), "oci registry")
}

func (c *ComponentVersionContainer) Update() error {
	err := c.Check()
	if err != nil {
		return err
	}
	if c.state.HasChanged() {
		desc := c.GetDescriptor()
		for i, r := range desc.Resources {
			s, err := c.evalLayer(r.Access)
			if err != nil {
				return err
			}
			if s != r.Access {
				desc.Resources[i].Access = s
			}
		}
		for i, r := range desc.Sources {
			s, err := c.evalLayer(r.Access)
			if err != nil {
				return err
			}
			if s != r.Access {
				desc.Sources[i].Access = s
			}
		}
		_, err = c.state.Update()
		if err != nil {
			return err
		}
		_, err = c.comp.namespace.AddArtefact(c.manifest, c.version)
	}
	return err
}

func (c *ComponentVersionContainer) evalLayer(s compdesc.AccessSpec) (compdesc.AccessSpec, error) {
	spec, err := c.GetContext().AccessSpecForSpec(s)
	if err != nil {
		return s, err
	}
	if a, ok := spec.(*localblob.AccessSpec); ok {
		if ok, _ := artdesc.IsDigest(a.LocalReference); !ok {
			return s, errors.ErrInvalid("digest", a.LocalReference)
		}
	}
	return s, nil
}

func (c *ComponentVersionContainer) GetDescriptor() *compdesc.ComponentDescriptor {
	return c.state.GetState().(*compdesc.ComponentDescriptor)
}

func (c *ComponentVersionContainer) GetBlobData(name string) (cpi.DataAccess, error) {
	return c.manifest.GetBlob(digest.Digest((name)))
}

func (c *ComponentVersionContainer) GetStorageContext(cv cpi.ComponentVersionAccess) cpi.StorageContext {
	return ocihdlr.New(c.comp.repo, cv, c.comp.repo.ocirepo.GetSpecification().GetKind(), c.comp.repo.ocirepo, c.comp.namespace, c.manifest)
}

func (c *ComponentVersionContainer) AddBlobFor(storagectx cpi.StorageContext, blob cpi.BlobAccess, refName string, global cpi.AccessSpec) (cpi.AccessSpec, error) {
	if blob == nil {
		return nil, errors.New("a resource has to be defined")
	}

	err := c.manifest.AddBlob(blob)
	if err != nil {
		return nil, err
	}
	err = storagectx.(*ocihdlr.StorageContext).AssureLayer(blob)
	if err != nil {
		return nil, err
	}
	return localblob.New(blob.Digest().String(), refName, blob.MimeType(), global), nil
}

// AssureGlobalRef provides a global manifest for a local OCI Artefact
func (c *ComponentVersionContainer) AssureGlobalRef(d digest.Digest, url, name string) (cpi.AccessSpec, error) {

	blob, err := c.manifest.GetBlob(d)
	if err != nil {
		return nil, err
	}
	var namespace oci.NamespaceAccess
	var version string
	var tag string
	if name == "" {
		namespace = c.comp.namespace
	} else {
		i := strings.LastIndex(name, ":")
		if i > 0 {
			version = name[i+1:]
			name = name[:i]
			tag = version
		}
		namespace, err = c.comp.repo.ocirepo.LookupNamespace(name)
		if err != nil {
			return nil, err
		}
	}
	set, err := artefactset.OpenFromBlob(accessobj.ACC_READONLY, blob)
	if err != nil {
		return nil, err
	}
	defer set.Close()
	digest := set.GetMain()
	if version == "" {
		version = digest.String()
	}
	art, err := set.GetArtefact(digest.String())
	if err != nil {
		return nil, err
	}
	err = artefactset.TransferArtefact(art, namespace, oci.AsTags(tag)...)
	if err != nil {
		return nil, err
	}

	ref := path.Join(url+namespace.GetNamespace()) + ":" + version

	global := ociartefact.New(ref)
	return global, nil
}
