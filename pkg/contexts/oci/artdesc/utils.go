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

	"github.com/opencontainers/go-digest"

	"github.com/open-component-model/ocm/pkg/common/accessio"
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

func IsDigest(version string) (bool, digest.Digest) {
	if strings.HasPrefix(version, "@") {
		return true, digest.Digest(version[1:])
	}
	if strings.Contains(version, ":") {
		return true, digest.Digest(version)
	}
	return false, ""
}

func ToContentMediaType(media string) string {
loop:
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
		case "yaml":
			fallthrough
		case "json":
			media = media[:last]
		default:
			break loop
		}
	}
	return media
}

func ToDescriptorMediaType(media string) string {
	return ToContentMediaType(media) + "+json"
}

func IsOCIMediaType(media string) bool {
	c := ToContentMediaType(media)
	for _, t := range ContentTypes() {
		if t == c {
			return true
		}
	}
	return false
}

func ContentTypes() []string {
	r := []string{}
	for _, t := range DescriptorTypes() {
		r = append(r, ToContentMediaType(t))
	}
	return r
}

func DescriptorTypes() []string {
	return []string{
		MediaTypeImageManifest,
		MediaTypeImageIndex,
		MediaTypeDockerSchema2Manifest,
		MediaTypeDockerSchema2ManifestList,
	}
}

func ArchiveBlobTypes() []string {
	r := []string{}
	for _, t := range ContentTypes() {
		t = ToContentMediaType(t)
		r = append(r, t+"+tar", t+"+tar+gzip")
	}
	return r
}

func ArtefactMimeType(cur, def string, legacy bool) string {
	if cur != "" {
		return cur
	}
	return MapArtefactMimeType(def, legacy)
}

func MapArtefactMimeType(mime string, legacy bool) string {
	if legacy {
		switch mime {
		case MediaTypeImageManifest:
			return MediaTypeDockerSchema2Manifest
		case MediaTypeImageIndex:
			return MediaTypeDockerSchema2ManifestList
		}
	} else {
		switch mime {
		case MediaTypeDockerSchema2Manifest:
			//return MediaTypeImageManifest
		case MediaTypeDockerSchema2ManifestList:
			//return MediaTypeImageIndex
		}
	}
	return mime
}

func MapArtefactBlobMimeType(blob accessio.BlobAccess, legacy bool) accessio.BlobAccess {
	mime := blob.MimeType()
	mapped := MapArtefactMimeType(mime, legacy)
	if mapped != mime {
		return accessio.BlobWithMimeType(mapped, blob)
	}
	return blob
}
