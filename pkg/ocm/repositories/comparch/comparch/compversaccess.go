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

package comparch

import (
	"io"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/ocm/accessmethods/localblob"
	"github.com/open-component-model/ocm/pkg/ocm/compdesc"
	metav1 "github.com/open-component-model/ocm/pkg/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/ocm/cpi"
)

type ComponentVersionAccess struct {
	lazy bool
	base ComponentVersionContainer
}

var _ cpi.ComponentVersionAccess = (*ComponentVersionAccess)(nil)

func NewComponentVersionAccess(container ComponentVersionContainer, lazy bool) *ComponentVersionAccess {
	s := &ComponentVersionAccess{
		lazy: lazy,
		base: container,
	}
	return s
}

func (a *ComponentVersionAccess) Close() error {
	return errors.ErrListf("closing archive").Add(a.base.Update(), a.base.Close()).Result()
}

func (a *ComponentVersionAccess) IsReadOnly() bool {
	return a.base.IsReadOnly()
}

func (a *ComponentVersionAccess) IsClosed() bool {
	return a.base.IsClosed()
}

func (a *ComponentVersionAccess) GetContext() cpi.Context {
	return a.base.GetContext()
}

func (a *ComponentVersionAccess) GetName() string {
	return a.base.GetDescriptor().GetName()
}

func (a *ComponentVersionAccess) GetVersion() string {
	return a.base.GetDescriptor().GetVersion()
}

func (a *ComponentVersionAccess) AddBlob(blob cpi.BlobAccess, refName string, global cpi.AccessSpec) (cpi.AccessSpec, error) {
	return a.base.AddBlob(blob, refName, global)
}

func (c *ComponentVersionAccess) AccessMethod(a cpi.AccessSpec) (cpi.AccessMethod, error) {
	if !a.IsLocal(c.base.GetContext()) {
		// fall back to original version
		return a.AccessMethod(c)
	}
	if a.GetKind() == localblob.Type || a.GetKind() == LocalFilesystemBlobType {
		a, err := c.base.GetContext().AccessSpecForSpec(a)
		if err != nil {
			return nil, err
		}
		return newLocalFilesystemBlobAccessMethod(a.(*localblob.AccessSpec), c)
	}
	return nil, errors.ErrNotSupported(errors.KIND_ACCESSMETHOD, a.GetType(), "component archive")
}

func (a *ComponentVersionAccess) GetDescriptor() *compdesc.ComponentDescriptor {
	return a.base.GetDescriptor()
}

func (a *ComponentVersionAccess) GetResource(id metav1.Identity) (cpi.ResourceAccess, error) {
	r, err := a.base.GetDescriptor().GetResourceByIdentity(id)
	if err != nil {
		return nil, err
	}
	return &ResourceAccess{
		BaseAccess: &BaseAccess{
			vers:   a,
			access: r.Access,
		},
		meta: r.ResourceMeta,
	}, nil
}

func (a *ComponentVersionAccess) GetResources() []cpi.ResourceAccess {
	result := []cpi.ResourceAccess{}
	for _, r := range a.GetDescriptor().Resources {
		result = append(result, &ResourceAccess{
			BaseAccess: &BaseAccess{
				vers:   a,
				access: r.Access,
			},
			meta: r.ResourceMeta,
		})
	}
	return result
}

func (a *ComponentVersionAccess) GetSource(id metav1.Identity) (cpi.SourceAccess, error) {
	r, err := a.base.GetDescriptor().GetSourceByIdentity(id)
	if err != nil {
		return nil, err
	}
	return &SourceAccess{
		BaseAccess: &BaseAccess{
			vers:   a,
			access: r.Access,
		},
		meta: r.SourceMeta,
	}, nil
}

func (a *ComponentVersionAccess) GetSources() []cpi.SourceAccess {
	result := []cpi.SourceAccess{}
	for _, r := range a.GetDescriptor().Sources {
		result = append(result, &SourceAccess{
			BaseAccess: &BaseAccess{
				vers:   a,
				access: r.Access,
			},
			meta: r.SourceMeta,
		})
	}
	return result
}

func (c *ComponentVersionAccess) getAccessSpec(acc compdesc.AccessSpec) (cpi.AccessSpec, error) {
	return c.GetContext().AccessSpecForSpec(acc)
}

func (c *ComponentVersionAccess) getAccessMethod(acc compdesc.AccessSpec) (cpi.AccessMethod, error) {
	spec, err := c.getAccessSpec(acc)
	if err != nil {
		return nil, err
	}
	if spec, err := c.AccessMethod(spec); err != nil {
		return nil, err
	} else {
		return spec, nil
	}
}

func (c *ComponentVersionAccess) checkAccessSpec(acc compdesc.AccessSpec) error {
	_, err := c.getAccessMethod(acc)
	return err
}

func (c *ComponentVersionAccess) SetResource(meta *cpi.ResourceMeta, acc compdesc.AccessSpec) error {
	if err := c.checkAccessSpec(acc); err != nil {
		return err
	}
	res := &compdesc.Resource{
		ResourceMeta: *meta.Copy(),
		Access:       acc,
	}

	if idx := c.GetDescriptor().GetResourceIndex(meta); idx == -1 {
		c.GetDescriptor().Resources = append(c.GetDescriptor().Resources, *res)
	} else {
		c.GetDescriptor().Resources[idx] = *res
	}
	if c.lazy {
		return nil
	}
	return c.base.Update()
}

func (c *ComponentVersionAccess) SetSource(meta *cpi.SourceMeta, acc compdesc.AccessSpec) error {
	if err := c.checkAccessSpec(acc); err != nil {
		if !errors.IsErrUnknown(err) {
			return err
		}
	}
	res := &compdesc.Source{
		SourceMeta: *meta.Copy(),
		Access:     acc,
	}

	if idx := c.GetDescriptor().GetSourceIndex(meta); idx == -1 {
		c.GetDescriptor().Sources = append(c.GetDescriptor().Sources, *res)
	} else {
		c.GetDescriptor().Sources[idx] = *res
	}
	if c.lazy {
		return nil
	}
	return c.base.Update()
}

// AddResource adds a blob resource to the current archive.
func (c *ComponentVersionAccess) SetResourceBlob(meta *cpi.ResourceMeta, blob cpi.BlobAccess, refName string, global cpi.AccessSpec) error {
	acc, err := c.AddBlob(blob, refName, global)
	if err != nil {
		return err
	}
	return c.SetResource(meta, acc)
}

func (c *ComponentVersionAccess) SetSourceBlob(meta *cpi.SourceMeta, blob cpi.BlobAccess, refName string, global cpi.AccessSpec) error {
	acc, err := c.AddBlob(blob, refName, global)
	if err != nil {
		return err
	}
	return c.SetSource(meta, acc)
}

////////////////////////////////////////////////////////////////////////////////

func (c *ComponentVersionAccess) SetReference(ref *cpi.ComponentReference) error {

	if idx := c.GetDescriptor().GetComponentReferenceIndex(*ref); idx == -1 {
		c.GetDescriptor().ComponentReferences = append(c.GetDescriptor().ComponentReferences, *ref)
	} else {
		c.GetDescriptor().ComponentReferences[idx] = *ref
	}
	if c.lazy {
		return nil
	}
	return c.base.Update()
}

////////////////////////////////////////////////////////////////////////////////

type BaseAccess struct {
	vers   *ComponentVersionAccess
	access compdesc.AccessSpec
}

func (r BaseAccess) Access() (cpi.AccessSpec, error) {
	return r.vers.getAccessSpec(r.access)
}

func (r BaseAccess) AccessMethod() (cpi.AccessMethod, error) {
	return r.vers.getAccessMethod(r.access)
}

func (r BaseAccess) Get() ([]byte, error) {
	return accessio.BlobData(r.AccessMethod())
}

func (r BaseAccess) Reader() (io.ReadCloser, error) {
	return accessio.BlobReader(r.AccessMethod())
}

func (r BaseAccess) MimeType() string {
	m, err := r.AccessMethod()
	if err != nil {
		return ""
	}
	return m.MimeType()
}

////////////////////////////////////////////////////////////////////////////////

type ResourceAccess struct {
	*BaseAccess
	meta cpi.ResourceMeta
}

var _ cpi.ResourceAccess = (*ResourceAccess)(nil)

func (r ResourceAccess) Meta() *cpi.ResourceMeta {
	return &r.meta
}

////////////////////////////////////////////////////////////////////////////////

type SourceAccess struct {
	*BaseAccess
	meta cpi.SourceMeta
}

var _ cpi.SourceAccess = (*SourceAccess)(nil)

func (r SourceAccess) Meta() *cpi.SourceMeta {
	return &r.meta
}
