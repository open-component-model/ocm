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

package docker

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/containers/image/v5/image"
	"github.com/containers/image/v5/manifest"
	"github.com/containers/image/v5/types"
	"github.com/opencontainers/go-digest"
	"github.com/sirupsen/logrus"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/contexts/oci/cpi"
)

// fakeSource implements required methods to call the manifest conversion.
type fakeSource struct {
	types.ImageSource
	art   cpi.BlobAccess
	blobs cpi.BlobSource
	ref   types.ImageReference
}

func (f *fakeSource) GetManifest(ctx context.Context, instanceDigest *digest.Digest) ([]byte, string, error) {
	if instanceDigest != nil {
		return nil, "", fmt.Errorf("manifest lists are not supported")
	}
	data, err := f.art.Get()
	if err != nil {
		return nil, "", err
	}
	return data, f.art.MimeType(), nil
}

func (f *fakeSource) GetBlob(ctx context.Context, bi types.BlobInfo, bc types.BlobInfoCache) (io.ReadCloser, int64, error) {
	_, blob, err := f.blobs.GetBlobData(bi.Digest)
	if err != nil {
		return nil, accessio.BLOB_UNKNOWN_SIZE, err
	}

	r, err := blob.Reader()
	return r, bi.Size, err
}

func (f *fakeSource) Reference() types.ImageReference {
	return f.ref
}

////////////////////////////////////////////////////////////////////////////////

type artBlobCache struct {
	access cpi.ArtefactAccess
}

var _ accessio.BlobCache = (*artBlobCache)(nil)

func ArtefactAsBlobCache(access cpi.ArtefactAccess) accessio.BlobCache {
	return &artBlobCache{access}
}

func (a *artBlobCache) Ref() error {
	return nil
}

func (a *artBlobCache) Unref() error {
	return nil
}

func (a *artBlobCache) GetBlobData(digest digest.Digest) (int64, accessio.DataAccess, error) {
	blob, err := a.access.GetBlob(digest)
	if err != nil {
		return -1, nil, err
	}
	return blob.Size(), blob, err
}

func (a *artBlobCache) AddBlob(blob accessio.BlobAccess) (int64, digest.Digest, error) {
	err := a.access.AddBlob(blob)
	if err != nil {
		return -1, "", err
	}
	return blob.Size(), blob.Digest(), err
}

func (c *artBlobCache) AddData(data accessio.DataAccess) (int64, digest.Digest, error) {
	return c.AddBlob(accessio.BlobAccessForDataAccess(accessio.BLOB_UNKNOWN_DIGEST, accessio.BLOB_UNKNOWN_SIZE, "", data))
}

////////////////////////////////////////////////////////////////////////////////

func blobSource(art cpi.Artefact, blobs accessio.BlobSource) (accessio.BlobSource, error) {
	var err error
	if blobs == nil {
		if t, ok := art.(cpi.ArtefactAccess); !ok {
			return nil, fmt.Errorf("blob source required")
		} else {
			blobs = ArtefactAsBlobCache(t)
		}
	} else {
		if t, ok := art.(cpi.ArtefactAccess); ok {
			blobs, err = accessio.NewCascadedBlobCacheForSource(blobs, ArtefactAsBlobCache(t))
			if err != nil {
				return nil, err
			}
		}
	}
	return blobs, nil
}

func Convert(art cpi.Artefact, blobs accessio.BlobSource, dst types.ImageDestination) (cpi.BlobAccess, error) {
	blobs, err := blobSource(art, blobs)
	if err != nil {
		return nil, err
	}
	artblob, err := art.Blob()
	if err != nil {
		return nil, err
	}
	ociImage := &fakeSource{
		art:   artblob,
		blobs: blobs,
		ref:   dst.Reference(),
	}

	m, err := art.Manifest()
	if err != nil {
		return nil, err
	}
	for i, l := range m.Layers {
		size, blob, err := blobs.GetBlobData(l.Digest)
		if err != nil {
			return nil, err
		}
		r, err := blob.Reader()
		if err != nil {
			return nil, err
		}
		defer r.Close()
		bi := types.BlobInfo{
			Digest:      l.Digest,
			Size:        size,
			URLs:        l.URLs,
			Annotations: l.Annotations,
			MediaType:   l.MediaType,
		}
		logrus.Infof("put blob  for layer %d", i)
		_, err = dst.PutBlob(dummyContext, r, bi, nil, false)
		if err != nil {
			return nil, err
		}
	}

	un := image.UnparsedInstance(ociImage, nil)
	img, err := image.FromUnparsedImage(dummyContext, nil, un)
	if err != nil {
		return nil, err
	}

	opts := types.ManifestUpdateOptions{
		ManifestMIMEType: manifest.DockerV2Schema2MediaType,
		InformationOnly: types.ManifestUpdateInformation{
			Destination: dst,
		},
	}

	img, err = img.UpdatedImage(dummyContext, opts)
	if err != nil {
		return nil, err
	}

	bi := img.ConfigInfo()
	blob, err := img.ConfigBlob(dummyContext)
	if err != nil {
		return nil, err
	}
	var reader io.ReadCloser
	if blob == nil {
		_, orig, err := blobs.GetBlobData(bi.Digest)
		if err != nil {
			return nil, err
		}
		reader, err = orig.Reader()
		if err != nil {
			return nil, err
		}
	} else {
		reader = io.NopCloser(bytes.NewReader(blob))
	}
	_, err = dst.PutBlob(dummyContext, reader, bi, nil, true)
	if err != nil {
		return nil, err
	}
	man, _, err := img.Manifest(dummyContext)
	if err != nil {
		return nil, err
	}

	return artblob, dst.PutManifest(dummyContext, man, nil)
}
