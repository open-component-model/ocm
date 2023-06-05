// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package blueprint

import (
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/open-component-model/ocm/pkg/contexts/oci/artdesc"
	"github.com/open-component-model/ocm/pkg/mime"

	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	registry "github.com/open-component-model/ocm/pkg/contexts/ocm/download"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/resourcetypes"
	"github.com/open-component-model/ocm/pkg/finalizer"
)

const TYPE = resourcetypes.BLUEPRINT
const LEGACY_TYPE = resourcetypes.BLUEPRINT_LEGACY

type Extractor func(access accessio.DataAccess, path string, fs vfs.FileSystem) error

var mediaTypeSet map[string]Extractor

type Handler struct{}

func init() {
	mediaTypeSet = map[string]Extractor{
		mime.MIME_TAR:     ExtractArchive,
		mime.MIME_TGZ:     ExtractArchive,
		mime.MIME_TGZ_ALT: ExtractArchive,
	}
	for _, t := range append(artdesc.ToArchiveMediaTypes(artdesc.MediaTypeImageManifest), artdesc.ToArchiveMediaTypes(artdesc.MediaTypeDockerSchema2Manifest)...) {
		mediaTypeSet[t] = ExtractArtifact
	}

	registry.Register(&Handler{}, registry.ForArtifactType(TYPE))
	registry.Register(&Handler{}, registry.ForArtifactType(LEGACY_TYPE))
}

func (h Handler) Download(p common.Printer, racc cpi.ResourceAccess, path string, fs vfs.FileSystem) (_ bool, _ string, err error) {
	var finalize finalizer.Finalizer
	defer finalize.FinalizeWithErrorPropagationf(&err, "downloading blueprint")

	meth, err := racc.AccessMethod()
	if err != nil {
		return false, "", err
	}
	finalize.Close(meth)

	ex := mediaTypeSet[meth.MimeType()]
	if ex == nil {
		return false, "", nil
	}

	err = ex(meth, path, fs)
	if err != nil {
		return false, "", err
	}
	return true, path, nil
}
