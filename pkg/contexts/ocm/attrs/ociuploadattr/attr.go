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

package ociuploadattr

import (
	"fmt"
	"sync"

	"github.com/open-component-model/ocm/pkg/contexts/datacontext"
	"github.com/open-component-model/ocm/pkg/contexts/oci"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/runtime"
)

const ATTR_KEY = "github.com/mandelsoft/ocm/ociuploadrepo"
const ATTR_SHORT = "ociuploadrepo"

func init() {
	datacontext.RegisterAttributeType(ATTR_KEY, AttributeType{}, ATTR_SHORT)
}

type AttributeType struct {
}

func (a AttributeType) Name() string {
	return ATTR_KEY
}

func (a AttributeType) Description() string {
	return `
*oci base repository ref*
Upload local OCI artefact blobs to a dedicated repository.
`
}

func (a AttributeType) Encode(v interface{}, marshaller runtime.Marshaler) ([]byte, error) {
	if _, ok := v.(*Attribute); !ok {
		return nil, fmt.Errorf("OCI Upload Attribute structure required")
	}
	return marshaller.Marshal(v)
}

func (a AttributeType) Decode(data []byte, unmarshaller runtime.Unmarshaler) (interface{}, error) {
	var value Attribute
	err := unmarshaller.Unmarshal(data, &value)
	if err == nil {
		return &value, nil
	}
	ref, err := oci.ParseRef(string(data))
	if err != nil {
		return nil, errors.ErrInvalidWrap(err, oci.KIND_OCI_REFERENCE, string(data))
	}
	if ref.Tag != nil || ref.Digest != nil {
		return nil, errors.ErrInvalidWrap(err, oci.KIND_OCI_REFERENCE, string(data))
	}
	return &Attribute{Ref: string(data)}, nil
}

////////////////////////////////////////////////////////////////////////////////

type Attribute struct {
	Ref string `json:"ociRef"`

	lock     sync.Mutex
	ref      *oci.RefSpec
	repo     oci.Repository
	baserepo string
}

func New(ref string) *Attribute {
	return &Attribute{Ref: ref}
}

func (a *Attribute) Close() error {
	a.lock.Lock()
	defer a.lock.Unlock()
	if a.repo != nil {
		defer func() {
			a.repo = nil
			a.baserepo = ""
		}()
		return a.repo.Close()
	}
	return nil
}

func (a *Attribute) GetInfo(ctx cpi.Context) (oci.Repository, string, string, error) {

	ref, err := oci.ParseRef(a.Ref)
	if err != nil {
		return nil, "", "", errors.ErrInvalidWrap(err, oci.KIND_OCI_REFERENCE, a.Ref)
	}
	if ref.Tag != nil || ref.Digest != nil {
		return nil, "", "", errors.ErrInvalidWrap(err, oci.KIND_OCI_REFERENCE, a.Ref)
	}

	a.lock.Lock()
	defer a.lock.Unlock()
	if a.ref == nil || ref != *a.ref {
		if a.repo != nil {
			a.repo.Close()
			a.repo = nil
		}

		spec, err := ctx.OCIContext().MapUniformRepositorySpec(&ref.UniformRepositorySpec)
		if err != nil {
			return nil, "", "", err
		}
		a.repo, err = ctx.OCIContext().RepositoryForSpec(spec)
		if err != nil {
			return nil, "", "", err
		}
		a.baserepo = ref.Repository
		a.ref = &ref
	}
	return a.repo, a.ref.UniformRepositorySpec.String(), a.baserepo, nil
}

////////////////////////////////////////////////////////////////////////////////

func Get(ctx datacontext.Context) *Attribute {
	a := ctx.GetAttributes().GetAttribute(ATTR_KEY)
	if a == nil {
		return nil
	}
	return a.(*Attribute)
}

func Set(ctx datacontext.Context, attr *Attribute) error {
	return ctx.GetAttributes().SetAttribute(ATTR_KEY, attr)
}
