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

package artdesc

import (
	"encoding/json"

	"github.com/opencontainers/go-digest"
	"github.com/opencontainers/image-spec/specs-go"
	ociv1 "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/open-component-model/ocm/pkg/common/accessio"
)

type Index ociv1.Index

var _ BlobDescriptorSource = (*Index)(nil)

func NewIndex() *Index {
	return &Index{
		Versioned:   specs.Versioned{SchemeVersion},
		MediaType:   MediaTypeImageIndex,
		Manifests:   nil,
		Annotations: nil,
	}
}

func (i *Index) IsValid() bool {
	return true
}

func (i *Index) GetBlobDescriptor(digest digest.Digest) *Descriptor {
	for _, m := range i.Manifests {
		if m.Digest == digest {
			return &m
		}
	}
	return nil
}

func (i *Index) MimeType() string {
	return ArtefactMimeType(i.MediaType, MediaTypeImageIndex, legacy)
}

func (i *Index) ToBlobAccess() (accessio.BlobAccess, error) {
	i.MediaType = i.MimeType()
	data, err := json.Marshal(i)
	if err != nil {
		return nil, err
	}
	return accessio.BlobAccessForData(i.MediaType, data), nil
}

func (i *Index) AddManifest(d *Descriptor) {
	i.Manifests = append(i.Manifests, *d)
}

////////////////////////////////////////////////////////////////////////////////

func DecodeIndex(data []byte) (*Index, error) {
	var d Index

	if err := json.Unmarshal(data, &d); err != nil {
		return nil, err
	}
	return &d, nil
}

func EncodeIndex(d *Index) ([]byte, error) {
	return json.Marshal(d)
}
