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
	"strings"

	"github.com/gardener/ocm/pkg/common/accessio"
)

func DefaultBlobDescriptor(blob accessio.BlobAccess) *Descriptor {
	return &Descriptor{
		MediaType:   blob.MimeType(),
		Digest:      blob.Digest(),
		Size:        blob.Size(),
		URLs:        nil,
		Annotations: nil,
		Platform:    nil,
	}
}

func IsDigest(ref string) bool {
	return strings.Index(ref, ":") >= 0
}

func ToContentMediaType(media string) string {
	for {
		last := strings.LastIndex(media, "+")
		if last < 0 {
			break
		}
		switch media[last+1:] {
		case "tar":
			fallthrough
		case "gzip":
			fallthrough
		case "json":
			media = media[:last]
		default:
			break
		}
	}
	return media
}

func ToDescriptorMediaType(media string) string {
	return ToContentMediaType(media) + "+json"
}

func IsOCIMediaType(media string) bool {
	switch ToDescriptorMediaType(media) {
	case MediaTypeImageIndex:
		fallthrough
	case MediaTypeImageManifest:
		return true
	default:
		return false
	}
}

func ContentTypes() []string {
	return []string{ToContentMediaType(MediaTypeImageManifest),
		ToContentMediaType(MediaTypeImageIndex),
	}
}

func ArchiveBlobTypes() []string {
	manifest := ToContentMediaType(MediaTypeImageManifest)
	index := ToContentMediaType(MediaTypeImageIndex)
	return []string{
		manifest + "+tar",
		manifest + "+tar+gzip",
		index + "+tar",
		index + "+tar+gzip",
	}
}
