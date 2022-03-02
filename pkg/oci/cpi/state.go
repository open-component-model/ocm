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
	"reflect"

	"github.com/gardener/ocm/pkg/common/accessobj"
	"github.com/gardener/ocm/pkg/oci/artdesc"
)

type ManifestStateHandler struct {
}

var _ accessobj.StateHandler = &ManifestStateHandler{}

func NewManifestStateHandler() accessobj.StateHandler {
	return &ManifestStateHandler{}
}

func (i ManifestStateHandler) Initial() interface{} {
	return artdesc.NewManifest()
}

func (i ManifestStateHandler) Encode(d interface{}) ([]byte, error) {
	return artdesc.EncodeManifest(d.(*artdesc.Manifest))
}

func (i ManifestStateHandler) Decode(data []byte) (interface{}, error) {
	return artdesc.DecodeManifest(data)
}

func (i ManifestStateHandler) Equivalent(a, b interface{}) bool {
	return reflect.DeepEqual(a, b)
}

////////////////////////////////////////////////////////////////////////////////

type IndexStateHandler struct {
}

var _ accessobj.StateHandler = &IndexStateHandler{}

func NewIndexStateHandler() accessobj.StateHandler {
	return &IndexStateHandler{}
}

func (i IndexStateHandler) Initial() interface{} {
	return artdesc.NewIndex()
}

func (i IndexStateHandler) Encode(d interface{}) ([]byte, error) {
	return artdesc.EncodeIndex(d.(*artdesc.Index))
}

func (i IndexStateHandler) Decode(data []byte) (interface{}, error) {
	return artdesc.DecodeIndex(data)
}

func (i IndexStateHandler) Equivalent(a, b interface{}) bool {
	return reflect.DeepEqual(a, b)
}

////////////////////////////////////////////////////////////////////////////////

type ArtefactStateHandler struct {
}

var _ accessobj.StateHandler = &ArtefactStateHandler{}

func NewArtefactStateHandler() accessobj.StateHandler {
	return &ArtefactStateHandler{}
}

func (i ArtefactStateHandler) Initial() interface{} {
	return artdesc.New()
}

func (i ArtefactStateHandler) Encode(d interface{}) ([]byte, error) {
	return artdesc.Encode(d.(*artdesc.Artefact))
}

func (i ArtefactStateHandler) Decode(data []byte) (interface{}, error) {
	return artdesc.Decode(data)
}

func (i ArtefactStateHandler) Equivalent(a, b interface{}) bool {
	return reflect.DeepEqual(a, b)
}
