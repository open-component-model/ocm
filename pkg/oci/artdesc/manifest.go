// Copyright 2020 Copyright (c) 2020 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package artdesc

import (
	"encoding/json"

	"github.com/gardener/ocm/pkg/oci/artdesc/helper"
	ociv1 "github.com/opencontainers/image-spec/specs-go/v1"
)

type Manifest = ociv1.Manifest
type Index = ociv1.Index

type ArtefactDescriptor struct {
	manifest *Manifest
	index    *Index
}

var _ json.Marshaler = &ArtefactDescriptor{}
var _ json.Unmarshaler = &ArtefactDescriptor{}

func (d *ArtefactDescriptor) IsManifest() bool {
	return d.manifest != nil
}

func (d *ArtefactDescriptor) IsIndex() bool {
	return d.index != nil
}

func (d *ArtefactDescriptor) Index() *Index {
	return d.index
}

func (d *ArtefactDescriptor) Manifest() *Manifest {
	return d.manifest
}

func (d ArtefactDescriptor) MarshalJSON() ([]byte, error) {
	if d.manifest != nil {
		d.manifest.MediaType = ociv1.MediaTypeImageManifest
		return json.Marshal(d.manifest)
	}
	if d.index != nil {
		d.manifest.MediaType = ociv1.MediaTypeImageIndex
		return json.Marshal(d.index)
	}
	return []byte("null"), nil
}

func (d ArtefactDescriptor) UnmarshalJSON(data []byte) error {
	var m helper.GenericDescriptor

	err := json.Unmarshal(data, &m)
	if err != nil {
		return err
	}

	err = m.Validate()
	if err != nil {
		return err
	}
	if m.IsManifest() {
		d.manifest = m.AsManifest()
		d.index = nil
	}
	d.index = m.AsIndex()
	d.manifest = nil
	return nil
}

func Decode(data []byte) (*ArtefactDescriptor, error) {
	var d ArtefactDescriptor

	if err := json.Unmarshal(data, &d); err != nil {
		return nil, err
	}
	return &d, nil
}

func Encode(d *ArtefactDescriptor) ([]byte, error) {
	return json.Marshal(d)
}
