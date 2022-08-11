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
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/ociartefact"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/ociuploadattr"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/errors"
)

func init() {
	for _, mime := range artdesc.ArchiveBlobTypes() {
		cpi.RegisterBlobHandler(NewArtefactHandler(), cpi.ForMimeType(mime), cpi.WithPrio(10))
	}
}

////////////////////////////////////////////////////////////////////////////////

// artefactHandler stores artefact blobs as OCIArtefacts
type artefactHandler struct {
}

func NewArtefactHandler() cpi.BlobHandler {
	return &artefactHandler{}
}

func (b *artefactHandler) StoreBlob(blob cpi.BlobAccess, hint string, global cpi.AccessSpec, ctx cpi.StorageContext) (cpi.AccessSpec, error) {
	attr := ociuploadattr.Get(ctx.GetContext())
	if attr == nil {
		return nil, nil
	}

	mediaType := blob.MimeType()
	if !artdesc.IsOCIMediaType(mediaType) || (!strings.HasSuffix(mediaType, "+tar") && !strings.HasSuffix(mediaType, "+tar+gzip")) {
		return nil, nil
	}

	repo, base, prefix, err := attr.GetInfo(ctx.GetContext())
	if err != nil {
		return nil, err
	}
	sep := ""
	if !ocireg.Is(repo.GetSpecification()) {
		// may be we should reject this, for non test scenario
		sep = "//"
	}
	var namespace oci.NamespaceAccess
	var version string
	var name string
	var tag string

	if hint == "" {
		name = path.Join(prefix, ctx.TargetComponentVersion().GetName())
	} else {
		i := strings.LastIndex(hint, ":")
		if i > 0 {
			version = hint[i:]
			name = path.Join(prefix, hint[:i])
			tag = version[1:] // remove colon
		} else {
			name = hint
		}
	}
	namespace, err = repo.LookupNamespace(name)
	if err != nil {
		return nil, errors.Wrapf(err, "lookup namespace %s in target repository %s", name, attr.Ref)
	}
	defer namespace.Close()

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
	if base != "" && sep != "" {
		ref = base + sep + namespace.GetNamespace() + version
	}
	var acc cpi.AccessSpec = ociartefact.New(ref)
	return acc, nil
}
