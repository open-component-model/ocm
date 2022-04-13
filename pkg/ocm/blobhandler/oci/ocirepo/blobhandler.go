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

	"github.com/gardener/ocm/pkg/common/accessobj"
	"github.com/gardener/ocm/pkg/oci"
	"github.com/gardener/ocm/pkg/oci/artdesc"
	artefactset2 "github.com/gardener/ocm/pkg/oci/repositories/artefactset"
	"github.com/gardener/ocm/pkg/oci/repositories/ocireg"
	"github.com/gardener/ocm/pkg/ocm/accessmethods/localblob"
	"github.com/gardener/ocm/pkg/ocm/accessmethods/ociblob"
	"github.com/gardener/ocm/pkg/ocm/accessmethods/ociregistry"
	storagecontext "github.com/gardener/ocm/pkg/ocm/blobhandler/oci"
	"github.com/gardener/ocm/pkg/ocm/core"
	"github.com/gardener/ocm/pkg/ocm/cpi"
	"github.com/gardener/ocm/pkg/ocm/repositories/genericocireg"
)

func init() {
	for _, mime := range artdesc.ContentTypes() {
		cpi.RegisterBlobHandler(NewArtefactHandler(nil), cpi.ForRepo(oci.CONTEXT_TYPE, ocireg.OCIRegistryRepositoryType),
			cpi.ForMimeType(mime))
	}
	cpi.RegisterBlobHandler(NewBlobHandler(nil), cpi.ForRepo(oci.CONTEXT_TYPE, ocireg.OCIRegistryRepositoryType))
}

////////////////////////////////////////////////////////////////////////////////

type BaseFunction func(ctx *storagecontext.StorageContext) string

// blobHandler is the default handling to store local blobs as local blobs but with an additional
// globally accessible OCIBlob access method
type blobHandler struct {
	base BaseFunction
}

func (h *blobHandler) GetBaseURL(ctx *storagecontext.StorageContext) string {
	if h.base != nil {
		return h.base(ctx)
	}
	return ctx.Repository.(*ocireg.Repository).GetBaseURL()
}

func NewBlobHandler(base BaseFunction) cpi.BlobHandler {
	return &blobHandler{base}
}

func (b *blobHandler) StoreBlob(repo cpi.Repository, blob cpi.BlobAccess, hint string, ctx cpi.StorageContext) (core.AccessSpec, error) {
	ocictx := ctx.(*storagecontext.StorageContext)
	base := b.GetBaseURL(ocictx)
	i := strings.LastIndex(hint, ":")
	if i > 0 {
		hint = hint[:i]
	}
	err := ocictx.Manifest.AddBlob(blob)
	if err != nil {
		return nil, err
	}
	err = ocictx.AssureLayer(blob)
	if err != nil {
		return nil, err
	}
	return localblob.New(blob.Digest().String(), "", blob.MimeType(), ociblob.New(path.Join(base, ocictx.Namespace.GetNamespace()), blob.Digest(), blob.MimeType(), blob.Size())), nil
}

////////////////////////////////////////////////////////////////////////////////

// artefactHandler stores artefact blobs as OCIArtefacts
type artefactHandler struct {
	blobHandler
}

func NewArtefactHandler(base BaseFunction) cpi.BlobHandler {
	return &artefactHandler{blobHandler{base}}
}

func (b *artefactHandler) StoreBlob(repo cpi.Repository, blob cpi.BlobAccess, hint string, ctx cpi.StorageContext) (core.AccessSpec, error) {
	mediaType := blob.MimeType()

	if !artdesc.IsOCIMediaType(mediaType) || (!strings.HasSuffix(mediaType, "+tar") && !strings.HasSuffix(mediaType, "+tar+gzip")) {
		return nil, nil
	}

	var namespace oci.NamespaceAccess
	var version string
	var name string
	var tag string
	var err error

	ocictx := ctx.(*storagecontext.StorageContext)
	base := b.GetBaseURL(ocictx)
	if hint == "" {
		namespace = ocictx.Namespace
	} else {
		spec := repo.GetSpecification().(*genericocireg.RepositorySpec)
		i := strings.LastIndex(hint, ":")
		if i > 0 {
			version = hint[i:]
			name = path.Join(spec.SubPath, hint[:i])
			tag = version[1:]
		} else {
			name = hint
		}
		namespace, err = ocictx.Repository.LookupNamespace(name)
		if err != nil {
			return nil, err
		}
	}

	set, err := artefactset2.OpenFromBlob(accessobj.ACC_READONLY, blob)
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

	err = artefactset2.TransferArtefact(art, namespace, oci.AsTags(tag)...)
	if err != nil {
		return nil, err
	}

	ref := path.Join(base, namespace.GetNamespace()) + version

	global := ociregistry.New(ref)
	return global, nil
}
