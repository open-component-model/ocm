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
	"reflect"

	"github.com/gardener/ocm/pkg/common"
	"github.com/gardener/ocm/pkg/common/accessobj"
	"github.com/gardener/ocm/pkg/errors"
	"github.com/gardener/ocm/pkg/oci"
	"github.com/gardener/ocm/pkg/oci/artdesc"
	"github.com/gardener/ocm/pkg/ocm/accessmethods"
	"github.com/gardener/ocm/pkg/ocm/compdesc"
	"github.com/gardener/ocm/pkg/ocm/cpi"
	"github.com/gardener/ocm/pkg/ocm/repositories/ctf/comparch"
	"github.com/opencontainers/go-digest"
	ociv1 "github.com/opencontainers/image-spec/specs-go/v1"
)

type ComponentVersion struct {
	container *ComponentVersionContainer
	*comparch.ComponentVersionAccess
}

var _ cpi.ComponentVersionAccess = (*ComponentVersion)(nil)

func NewComponentVersionAccess(mode accessobj.AccessMode, comp *ComponentAccess, version string, access oci.ManifestAccess) (*ComponentVersion, error) {
	c, err := NewComponentVersionContainer(mode, comp, version, access)
	if err != nil {
		return nil, err
	}
	return &ComponentVersion{
		container:              c,
		ComponentVersionAccess: comparch.NewComponentVersionAccess(c),
	}, nil
}

////////////////////////////////////////////////////////////////////////////////

type ComponentVersionContainer struct {
	comp    *ComponentAccess
	version string
	access  oci.ManifestAccess
	state   accessobj.State
}

var _ comparch.ComponentVersionContainer = (*ComponentVersionContainer)(nil)

func NewComponentVersionContainer(mode accessobj.AccessMode, comp *ComponentAccess, version string, access oci.ManifestAccess) (*ComponentVersionContainer, error) {
	state, err := NewState(mode, comp.name, version, access)
	if err != nil {
		return nil, err
	}
	return &ComponentVersionContainer{
		comp:    comp,
		version: version,
		access:  access,
		state:   state,
	}, nil
}

func (c *ComponentVersionContainer) GetContext() cpi.Context {
	return c.comp.GetContext()
}

func (c *ComponentVersionContainer) IsReadOnly() bool {
	return c.state.IsReadOnly()
}

func (c *ComponentVersionContainer) IsClosed() bool {
	return c.access == nil
}

func (c *ComponentVersionContainer) Update() error {
	_, err := c.state.Update()
	if err != nil {
		return err
	}
	desc := c.GetDescriptor()
	for _, r := range desc.Resources {
		c.evalLayer(r.Access)
	}
	for _, r := range desc.Sources {
		c.evalLayer(r.Access)
	}
	_, err = c.comp.namespace.AddTaggedArtefact(c.access, c.version)
	if err != nil {
		return err
	}
	return nil
}

func (c *ComponentVersionContainer) evalLayer(spec compdesc.AccessSpec) error {
	spec, err := c.GetContext().AccessSpecForSpec(spec)
	if err != nil {
		return err
	}
	if a, ok := spec.(*accessmethods.LocalBlobAccessSpec); ok {
		if !artdesc.IsDigest(a.LocalReference) {
			return errors.ErrInvalid("digest", a.LocalReference)
		}
		desc := c.access.GetDescriptor()
		for _, l := range desc.Layers {
			if l.Digest == digest.Digest(a.LocalReference) {
				if artdesc.IsOCIMediaType(l.MediaType) && c.comp.repo.ocirepo.SupportsDistributionSpec() {
					return c.assureGlobalRef(l.Digest, a.ReferenceName)
				}
			}
		}
		return errors.ErrUnknown("localReference", a.LocalReference)
	}
	return nil
}

func (c *ComponentVersionContainer) assureLayer(blob cpi.BlobAccess) error {
	d := artdesc.DefaultBlobDescriptor(blob)
	desc := c.access.GetDescriptor()

	found := -1
	for i, l := range desc.Layers {
		if reflect.DeepEqual(&l, *d) {
			return nil
		}
		if l.Digest == blob.Digest() {
			found = i
		}
	}
	if found > 0 { // ignore layer 0 used for component descriptor
		desc.Layers[found] = *d
	} else {
		if len(desc.Layers) == 0 {
			// fake descriptor layer
			desc.Layers = append(desc.Layers, ociv1.Descriptor{MediaType: ComponentDescriptorConfigMimeType})
		}
		desc.Layers = append(desc.Layers, *d)
	}
	return nil
}

func (c *ComponentVersionContainer) GetDescriptor() *compdesc.ComponentDescriptor {
	return c.state.GetState().(*compdesc.ComponentDescriptor)
}

func (c *ComponentVersionContainer) GetBlobData(name string) (cpi.DataAccess, error) {
	return c.access.GetBlob(digest.Digest((name)))
}

func (c *ComponentVersionContainer) AddBlob(blob cpi.BlobAccess, refName string) (cpi.AccessSpec, error) {
	if blob == nil {
		return nil, errors.New("a resource has to be defined")
	}
	err := c.access.AddBlob(blob)
	if err != nil {
		return nil, err
	}
	err = c.assureLayer(blob)
	if err != nil {
		return nil, err
	}
	return accessmethods.NewLocalBlobAccessSpecV1(common.DigestToFileName(blob.Digest()), refName, blob.MimeType()), nil
}

// assureGlobalRef provides a global access for a local OCI Artefact
func (c *ComponentVersionContainer) assureGlobalRef(d digest.Digest, name string) error {
	return nil
}
