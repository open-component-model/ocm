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

package support

import (
	"fmt"
	"strconv"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/errors"
)

type ComponentVersionAccess struct {
	view accessio.CloserView // handle close and refs
	*componentVersionAccessImpl
}

// implemented by view
// the rest is directly taken from the artefact set implementation

func (s *ComponentVersionAccess) Close() error {
	err := s.base.Update()
	if err != nil {
		s.view.Close()
		return err
	}
	return s.view.Close()
}

func (s *ComponentVersionAccess) IsClosed() bool {
	return s.view.IsClosed()
}

////////////////////////////////////////////////////////////////////////////////

type componentVersionAccessImpl struct {
	refs accessio.ReferencableCloser
	lazy bool
	base ComponentVersionContainer
}

var _ cpi.ComponentVersionAccess = (*ComponentVersionAccess)(nil)

func NewComponentVersionAccess(container ComponentVersionContainer, lazy bool) (*ComponentVersionAccess, error) {
	s := &componentVersionAccessImpl{
		lazy: lazy,
		base: container,
	}
	s.refs = accessio.NewRefCloser(s, true)
	return s.View(true)
}

func (a *componentVersionAccessImpl) View(main ...bool) (*ComponentVersionAccess, error) {
	v, err := a.refs.View(main...)
	if err != nil {
		return nil, err
	}
	return &ComponentVersionAccess{view: v, componentVersionAccessImpl: a}, nil
}

func (a *componentVersionAccessImpl) Close() error {
	return errors.ErrListf("closing access").Add(a.base.Update(), a.base.Close()).Result()
}

func (c *componentVersionAccessImpl) Repository() cpi.Repository {
	return c.base.Repository()
}

func (a *componentVersionAccessImpl) IsReadOnly() bool {
	return a.base.IsReadOnly()
}

func (a *componentVersionAccessImpl) IsClosed() bool {
	return a.base.IsClosed()
}

func (a *componentVersionAccessImpl) GetContext() cpi.Context {
	return a.base.GetContext()
}

func (a *componentVersionAccessImpl) GetName() string {
	descriptor, err := a.base.GetDescriptor()
	if err != nil {
		return ""
	}
	return descriptor.GetName()
}

func (a *componentVersionAccessImpl) GetVersion() string {
	descriptor, err := a.base.GetDescriptor()
	if err != nil {
		return ""
	}
	return descriptor.GetVersion()
}

func (a *componentVersionAccessImpl) AddBlob(blob cpi.BlobAccess, refName string, global cpi.AccessSpec) (cpi.AccessSpec, error) {
	if blob == nil {
		return nil, errors.New("a resource has to be defined")
	}
	storagectx := a.base.GetStorageContext(a)
	h := a.GetContext().BlobHandlers().GetHandler(storagectx.GetImplementationRepositoryType(), blob.MimeType())
	if h != nil {
		acc, err := h.StoreBlob(blob, refName, nil, storagectx)
		if err != nil {
			return nil, err
		}
		if acc != nil {
			return acc, nil
		}
	}
	return a.base.AddBlobFor(storagectx, blob, refName, global)
}

func (c *componentVersionAccessImpl) AccessMethod(a cpi.AccessSpec) (cpi.AccessMethod, error) {
	if !a.IsLocal(c.base.GetContext()) {
		// fall back to original version
		return a.AccessMethod(c)
	}
	return c.base.AccessMethod(a)
}

func (a *componentVersionAccessImpl) GetDescriptor() (*compdesc.ComponentDescriptor, error) {
	return a.base.GetDescriptor()
}

func (a *componentVersionAccessImpl) GetResource(id metav1.Identity) (cpi.ResourceAccess, error) {
	descriptor, err := a.base.GetDescriptor()
	if err != nil {
		return nil, fmt.Errorf("failed to get descriptor while getting resource: %w", err)
	}
	r, err := descriptor.GetResourceByIdentity(id)
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

func (a *componentVersionAccessImpl) GetResourceByIndex(i int) (cpi.ResourceAccess, error) {
	descriptor, err := a.base.GetDescriptor()
	if err != nil {
		return nil, fmt.Errorf("failed to get descriptor while getting resource by index: %w", err)
	}
	if i < 0 || i > len(descriptor.Resources) {
		return nil, errors.ErrInvalid("resource index", strconv.Itoa(i))
	}
	getDescriptor, err := a.base.GetDescriptor()
	if err != nil {
		return nil, fmt.Errorf("failed to get descriptor while getting resource: %w", err)
	}
	r := getDescriptor.Resources[i]
	return &ResourceAccess{
		BaseAccess: &BaseAccess{
			vers:   a,
			access: r.Access,
		},
		meta: r.ResourceMeta,
	}, nil
}

func (a *componentVersionAccessImpl) GetResources() []cpi.ResourceAccess {
	result := []cpi.ResourceAccess{}
	descriptor, err := a.GetDescriptor()
	if err != nil {
		fmt.Printf("failed to get descriptor: %s", err)
		return result
	}
	for _, r := range descriptor.Resources {
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

func (a *componentVersionAccessImpl) GetSource(id metav1.Identity) (cpi.SourceAccess, error) {
	descriptor, err := a.base.GetDescriptor()
	if err != nil {
		return nil, fmt.Errorf("failed to get descriptor while getting source: %w", err)
	}
	r, err := descriptor.GetSourceByIdentity(id)
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

func (a *componentVersionAccessImpl) GetSourceByIndex(i int) (cpi.SourceAccess, error) {
	descriptor, err := a.base.GetDescriptor()
	if err != nil {
		return nil, fmt.Errorf("failed to get descriptor while getting source by index: %w", err)
	}
	if i < 0 || i > len(descriptor.Sources) {
		return nil, errors.ErrInvalid("source index", strconv.Itoa(i))
	}
	getDescriptor, err := a.base.GetDescriptor()
	if err != nil {
		return nil, fmt.Errorf("failed to get descriptor while getting source by index: %w", err)
	}
	r := getDescriptor.Sources[i]
	return &SourceAccess{
		BaseAccess: &BaseAccess{
			vers:   a,
			access: r.Access,
		},
		meta: r.SourceMeta,
	}, nil
}

func (a *componentVersionAccessImpl) GetSources() []cpi.SourceAccess {
	result := []cpi.SourceAccess{}
	descriptor, err := a.GetDescriptor()
	if err != nil {
		fmt.Println("failed to get descriptor for sources: ", err)
		return result
	}
	for _, r := range descriptor.Sources {
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

func (a *componentVersionAccessImpl) GetReference(id metav1.Identity) (cpi.ComponentReference, error) {
	descriptor, err := a.base.GetDescriptor()
	if err != nil {
		return cpi.ComponentReference{}, fmt.Errorf("failed to get descriptor when getting reference: %w", err)
	}
	return descriptor.GetReferenceByIdentity(id)
}

func (a *componentVersionAccessImpl) GetReferenceByIndex(i int) (cpi.ComponentReference, error) {
	descriptor, err := a.base.GetDescriptor()
	if err != nil {
		return cpi.ComponentReference{}, fmt.Errorf("failed to get descriptor for get reference by index: %w", err)
	}
	if i < 0 || i > len(descriptor.References) {
		return cpi.ComponentReference{}, errors.ErrInvalid("reference index", strconv.Itoa(i))
	}
	return descriptor.References[i], nil
}

func (c *componentVersionAccessImpl) getAccessSpec(acc compdesc.AccessSpec) (cpi.AccessSpec, error) {
	return c.GetContext().AccessSpecForSpec(acc)
}

func (c *componentVersionAccessImpl) getAccessMethod(acc compdesc.AccessSpec) (cpi.AccessMethod, error) {
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

func (c *componentVersionAccessImpl) AdjustResourceAccess(meta *cpi.ResourceMeta, acc compdesc.AccessSpec) error {
	if err := c.checkAccessSpec(acc); err != nil {
		return err
	}

	cd, err := c.GetDescriptor()
	if err != nil {
		return fmt.Errorf("failed to get descriptor while adjusting resource access: %w", err)
	}
	if idx := cd.GetResourceIndex(meta); idx == -1 {
		return errors.ErrUnknown(cpi.KIND_RESOURCE, meta.GetIdentity(cd.Resources).String())
	} else {
		cd.Resources[idx].Access = acc
	}
	if c.lazy {
		return nil
	}
	return c.base.Update()
}

func (c *componentVersionAccessImpl) checkAccessSpec(acc compdesc.AccessSpec) error {
	_, err := c.getAccessMethod(acc)
	return err
}

func (c *componentVersionAccessImpl) SetResource(meta *cpi.ResourceMeta, acc compdesc.AccessSpec) error {
	if err := c.checkAccessSpec(acc); err != nil {
		return err
	}
	res := &compdesc.Resource{
		ResourceMeta: *meta.Copy(),
		Access:       acc,
	}

	if res.Relation == metav1.LocalRelation {
		switch res.Version {
		case "":
			res.Version = c.GetVersion()
		case c.GetVersion():
		default:
			return errors.ErrInvalid("resource version", res.Version)
		}
	}

	cd, err := c.GetDescriptor()
	if err != nil {
		return fmt.Errorf("failed to get descriptor while settin resources: %w", err)
	}
	if idx := cd.GetResourceIndex(meta); idx == -1 {
		cd.Resources = append(cd.Resources, *res)
		cd.Signatures = nil
	} else {
		if !cd.Resources[idx].ResourceMeta.HashEqual(&res.ResourceMeta) {
			cd.Signatures = nil
		}
		cd.Resources[idx] = *res
	}
	if c.lazy {
		return nil
	}
	return c.base.Update()
}

func (c *componentVersionAccessImpl) SetSource(meta *cpi.SourceMeta, acc compdesc.AccessSpec) error {
	if err := c.checkAccessSpec(acc); err != nil {
		if !errors.IsErrUnknown(err) {
			return err
		}
	}
	res := &compdesc.Source{
		SourceMeta: *meta.Copy(),
		Access:     acc,
	}

	switch res.Version {
	case "":
		res.Version = c.GetVersion()
	}

	descriptor, err := c.GetDescriptor()
	if err != nil {
		return err
	}
	if idx := descriptor.GetSourceIndex(meta); idx == -1 {
		descriptor.Sources = append(descriptor.Sources, *res)
	} else {
		descriptor.Sources[idx] = *res
	}
	if c.lazy {
		return nil
	}
	return c.base.Update()
}

// AddResource adds a blob resource to the current archive.
func (c *componentVersionAccessImpl) SetResourceBlob(meta *cpi.ResourceMeta, blob cpi.BlobAccess, refName string, global cpi.AccessSpec) error {
	acc, err := c.AddBlob(blob, refName, global)
	if err != nil {
		return err
	}
	return c.SetResource(meta, acc)
}

func (c *componentVersionAccessImpl) SetSourceBlob(meta *cpi.SourceMeta, blob cpi.BlobAccess, refName string, global cpi.AccessSpec) error {
	acc, err := c.AddBlob(blob, refName, global)
	if err != nil {
		return err
	}
	return c.SetSource(meta, acc)
}

////////////////////////////////////////////////////////////////////////////////

func (c *componentVersionAccessImpl) SetReference(ref *cpi.ComponentReference) error {

	descriptor, err := c.GetDescriptor()
	if err != nil {
		return err
	}
	if idx := descriptor.GetComponentReferenceIndex(*ref); idx == -1 {
		descriptor.References = append(descriptor.References, *ref)
	} else {
		descriptor.References[idx] = *ref
	}
	if c.lazy {
		return nil
	}
	return c.base.Update()
}

////////////////////////////////////////////////////////////////////////////////

type BaseAccess struct {
	vers   *componentVersionAccessImpl
	access compdesc.AccessSpec
}

func (r *BaseAccess) Access() (cpi.AccessSpec, error) {
	return r.vers.getAccessSpec(r.access)
}

func (r *BaseAccess) AccessMethod() (cpi.AccessMethod, error) {
	return r.vers.getAccessMethod(r.access)
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
