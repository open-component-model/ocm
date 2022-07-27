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

package cpi

import (
	"strconv"

	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/core"
	"github.com/open-component-model/ocm/pkg/errors"
)

type DummyComponentVersionAccess struct {
	Context Context
}

var _ ComponentVersionAccess = (*DummyComponentVersionAccess)(nil)

func (d *DummyComponentVersionAccess) GetContext() Context {
	return d.Context
}

func (c *DummyComponentVersionAccess) Repository() Repository {
	panic("implement me")
}

func (d *DummyComponentVersionAccess) GetName() string {
	panic("implement me")
}

func (d *DummyComponentVersionAccess) GetVersion() string {
	panic("implement me")
}

func (d *DummyComponentVersionAccess) GetDescriptor() *compdesc.ComponentDescriptor {
	panic("implement me")
}

func (d *DummyComponentVersionAccess) GetResources() []ResourceAccess {
	return nil
}

func (d *DummyComponentVersionAccess) GetResource(meta metav1.Identity) (ResourceAccess, error) {
	return nil, errors.ErrNotFound("resource", meta.String())
}

func (d *DummyComponentVersionAccess) GetResourceByIndex(i int) (ResourceAccess, error) {
	return nil, errors.ErrInvalid("resource index", strconv.Itoa(i))
}

func (d *DummyComponentVersionAccess) GetSources() []SourceAccess {
	panic("implement me")
}

func (d *DummyComponentVersionAccess) GetSource(meta metav1.Identity) (SourceAccess, error) {
	panic("implement me")
}

func (d *DummyComponentVersionAccess) GetSourceByIndex(i int) (SourceAccess, error) {
	return nil, errors.ErrInvalid("source index", strconv.Itoa(i))
}

func (d *DummyComponentVersionAccess) GetReference(meta metav1.Identity) (ComponentReference, error) {
	return ComponentReference{}, errors.ErrNotFound("reference", meta.String())
}

func (d *DummyComponentVersionAccess) GetReferenceByIndex(i int) (ComponentReference, error) {
	return ComponentReference{}, errors.ErrInvalid("reference index", strconv.Itoa(i))
}

func (d *DummyComponentVersionAccess) AccessMethod(spec AccessSpec) (AccessMethod, error) {
	panic("implement me")
}

func (d *DummyComponentVersionAccess) AddBlob(blob BlobAccess, refName string, global AccessSpec) (AccessSpec, error) {
	panic("implement me")
}

func (d *DummyComponentVersionAccess) SetResourceBlob(meta *ResourceMeta, blob BlobAccess, refname string, global AccessSpec) error {
	panic("implement me")
}

func (d *DummyComponentVersionAccess) AdjustResourceAccess(meta *core.ResourceMeta, acc compdesc.AccessSpec) error {
	panic("implement me")
}

func (d *DummyComponentVersionAccess) SetResource(meta *ResourceMeta, spec compdesc.AccessSpec) error {
	panic("implement me")
}

func (d *DummyComponentVersionAccess) SetSourceBlob(meta *SourceMeta, blob BlobAccess, refname string, global AccessSpec) error {
	panic("implement me")
}

func (d *DummyComponentVersionAccess) SetSource(meta *SourceMeta, spec compdesc.AccessSpec) error {
	panic("implement me")
}

func (d *DummyComponentVersionAccess) SetReference(ref *ComponentReference) error {
	panic("implement me")
}

func (d *DummyComponentVersionAccess) Close() error {
	return nil
}
