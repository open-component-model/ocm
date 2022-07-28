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

package ocirepo

import (
	"path"
	"strings"

	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/oci"
	"github.com/open-component-model/ocm/pkg/contexts/oci/artdesc"
	"github.com/open-component-model/ocm/pkg/contexts/oci/repositories/artefactset"
	"github.com/open-component-model/ocm/pkg/contexts/oci/repositories/ocireg"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/localblob"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/localociblob"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/ociartefact"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/ociblob"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/compatattr"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/keepblobattr"
	storagecontext "github.com/open-component-model/ocm/pkg/contexts/ocm/blobhandler/oci"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/genericocireg"
)

func init() {
	for _, mime := range artdesc.ArchiveBlobTypes() {
		cpi.RegisterBlobHandler(NewArtefactHandler(OCIRegBaseFunction), cpi.ForRepo(oci.CONTEXT_TYPE, ocireg.RepositoryType),
			cpi.ForMimeType(mime))
		cpi.RegisterBlobHandler(NewArtefactHandler(OCIRegBaseFunction), cpi.ForRepo(oci.CONTEXT_TYPE, ocireg.LegacyRepositoryType),
			cpi.ForMimeType(mime))
		cpi.RegisterBlobHandler(NewArtefactHandler(OCIRegBaseFunction), cpi.ForRepo(oci.CONTEXT_TYPE, ocireg.ShortRepositoryType),
			cpi.ForMimeType(mime))
	}
	cpi.RegisterBlobHandler(NewBlobHandler(OCIRegBaseFunction), cpi.ForRepo(oci.CONTEXT_TYPE, ocireg.RepositoryType))
	cpi.RegisterBlobHandler(NewBlobHandler(OCIRegBaseFunction), cpi.ForRepo(oci.CONTEXT_TYPE, ocireg.ShortRepositoryType))
}

////////////////////////////////////////////////////////////////////////////////

type BaseFunction func(ctx *storagecontext.StorageContext) string

func OCIRegBaseFunction(ctx *storagecontext.StorageContext) string {
	return ctx.Repository.(*ocireg.Repository).GetBaseURL()
}

// blobHandler is the default handling to store local blobs as local blobs but with an additional
// globally accessible OCIBlob access method
type blobHandler struct {
	base BaseFunction
}

func (h *blobHandler) GetBaseURL(ctx *storagecontext.StorageContext) string {
	if h.base == nil {
		return ""
	}
	return h.base(ctx)
}

func NewBlobHandler(base BaseFunction) cpi.BlobHandler {
	return &blobHandler{base}
}

func (b *blobHandler) StoreBlob(blob cpi.BlobAccess, hint string, global cpi.AccessSpec, ctx cpi.StorageContext) (cpi.AccessSpec, error) {
	ocictx := ctx.(*storagecontext.StorageContext)

	err := ocictx.Manifest.AddBlob(blob)
	if err != nil {
		return nil, err
	}
	err = ocictx.AssureLayer(blob)
	if err != nil {
		return nil, err
	}
	if compatattr.Get(ctx.GetContext()) {
		return localociblob.New(blob.Digest()), nil
	} else {
		if global == nil {
			base := b.GetBaseURL(ocictx)
			if base != "" {
				global = ociblob.New(path.Join(base, ocictx.Namespace.GetNamespace()), blob.Digest(), blob.MimeType(), blob.Size())
			}
		}
		return localblob.New(blob.Digest().String(), "", blob.MimeType(), global), nil
	}
}

////////////////////////////////////////////////////////////////////////////////

// artefactHandler stores artefact blobs as OCIArtefacts
type artefactHandler struct {
	blobHandler
}

func NewArtefactHandler(base BaseFunction) cpi.BlobHandler {
	return &artefactHandler{blobHandler{base}}
}

func (b *artefactHandler) StoreBlob(blob cpi.BlobAccess, hint string, global cpi.AccessSpec, ctx cpi.StorageContext) (cpi.AccessSpec, error) {
	mediaType := blob.MimeType()

	if !artdesc.IsOCIMediaType(mediaType) || (!strings.HasSuffix(mediaType, "+tar") && !strings.HasSuffix(mediaType, "+tar+gzip")) {
		return nil, nil
	}

	var namespace oci.NamespaceAccess
	var version string
	var name string
	var tag string
	var err error

	keep := keepblobattr.Get(ctx.GetContext())

	ocictx := ctx.(*storagecontext.StorageContext)
	base := b.GetBaseURL(ocictx)
	if hint == "" {
		namespace = ocictx.Namespace
	} else {
		spec := ctx.TargetComponentRepository().GetSpecification().(*genericocireg.RepositorySpec)
		i := strings.LastIndex(hint, ":")
		if i > 0 {
			version = hint[i:]
			name = path.Join(spec.SubPath, hint[:i])
			tag = version[1:] // remove colon
		} else {
			name = hint
		}
		namespace, err = ocictx.Repository.LookupNamespace(name)
		if err != nil {
			return nil, err
		}
		defer namespace.Close()
	}

	set, err := artefactset.OpenFromBlob(accessobj.ACC_READONLY, blob)
	if err != nil {
		return nil, err
	}
	defer set.Close()
	digest := set.GetMain()
	if version == "" {
		version = "@" + digest.String()
	}
	art, err := set.GetArtefact(digest.String())
	if err != nil {
		return nil, err
	}

	err = artefactset.TransferArtefact(art, namespace, oci.AsTags(tag)...)
	if err != nil {
		return nil, err
	}

	ref := path.Join(base, namespace.GetNamespace()) + version
	var acc cpi.AccessSpec = ociartefact.New(ref)

	if keep {
		err := ocictx.Manifest.AddBlob(blob)
		if err != nil {
			return nil, err
		}
		err = ocictx.AssureLayer(blob)
		if err != nil {
			return nil, err
		}
		acc = localblob.New(blob.Digest().String(), hint, blob.MimeType(), acc)
	}
	return acc, nil
}
