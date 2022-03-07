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
	"github.com/gardener/ocm/pkg/oci/cpi"
	"github.com/opencontainers/go-digest"
)

// fakeSource implements required methods to call the manifest conversion
type fakeSource struct {
	types.ImageSource
	art   cpi.Artefact
	blobs cpi.BlobSource
	ref   types.ImageReference
}

func (f *fakeSource) GetManifest(ctx context.Context, instanceDigest *digest.Digest) ([]byte, string, error) {
	if instanceDigest != nil {
		return nil, "", fmt.Errorf("manifest lists are not supported")
	}
	blob, err := f.art.Blob()
	if err != nil {
		return nil, "", err
	}
	data, err := blob.Get()
	if err != nil {
		return nil, "", err
	}
	return data, blob.MimeType(), nil
}

func (f *fakeSource) GetBlob(ctx context.Context, bi types.BlobInfo, bc types.BlobInfoCache) (io.ReadCloser, int64, error) {
	blob, err := f.blobs.GetBlobData(bi.Digest)
	if err != nil {
		return nil, -1, err
	}
	r, err := blob.Reader()
	return r, bi.Size, err
}

func (f *fakeSource) Reference() types.ImageReference {
	return f.ref
}

////////////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////////////

func Convert(art cpi.Artefact, blobs cpi.BlobSource, dst types.ImageDestination) error {

	if blobs == nil {
		if t, ok := art.(cpi.BlobSource); !ok {
			return fmt.Errorf("blob source required")
		} else {
			blobs = t
		}
	}
	ociImage := &fakeSource{
		art:   art,
		blobs: blobs,
		ref:   dst.Reference(),
	}

	m, err := art.Manifest()
	if err != nil {
		return err
	}
	for i, l := range m.Layers {
		blob, err := blobs.GetBlobData(l.Digest)
		if err != nil {
			return err
		}
		r, err := blob.Reader()
		if err != nil {
			return err
		}
		defer r.Close()
		info := types.BlobInfo{
			Digest:      l.Digest,
			Size:        -1,
			URLs:        l.URLs,
			Annotations: l.Annotations,
			MediaType:   l.MediaType,
		}
		fmt.Printf("put blob  for layer %d\n", i)
		info, err = dst.PutBlob(dummyContext, r, info, nil, false)
		if err != nil {
			return err
		}
	}

	un := image.UnparsedInstance(ociImage, nil)
	img, err := image.FromUnparsedImage(dummyContext, nil, un)
	if err != nil {
		return err
	}

	opts := types.ManifestUpdateOptions{
		ManifestMIMEType: manifest.DockerV2Schema2MediaType,
		InformationOnly: types.ManifestUpdateInformation{
			Destination: dst,
		},
	}

	img, err = img.UpdatedImage(dummyContext, opts)
	if err != nil {
		return err
	}

	bi := img.ConfigInfo()
	blob, err := img.ConfigBlob(dummyContext)
	if err != nil {
		return err
	}
	var reader io.ReadCloser
	if blob == nil {
		orig, err := blobs.GetBlobData(bi.Digest)
		if err != nil {
			return err
		}
		reader, err = orig.Reader()
		if err != nil {
			return err
		}
	} else {
		reader = io.NopCloser(bytes.NewReader(blob))
	}
	bi, err = dst.PutBlob(dummyContext, reader, bi, nil, true)

	man, _, err := img.Manifest(dummyContext)
	if err != nil {
		return err
	}

	return dst.PutManifest(dummyContext, man, nil)
}
