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
	"github.com/opencontainers/go-digest"
)

func GetBlobDescriptorFromManifest(digest digest.Digest, manifest *Manifest) *Descriptor {
	if manifest.Config.Digest == digest {
		d := manifest.Config
		return &d
	}
	for _, l := range manifest.Layers {
		if l.Digest == digest {
			return &l
		}
	}
	return nil
}

func GetBlobDescriptorFromIndex(digest digest.Digest, index *Index) *Descriptor {
	for _, m := range index.Manifests {
		if m.Digest == digest {
			return &m
		}
	}
	return nil
}
