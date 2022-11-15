// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package ocirepo

import (
	"fmt"
	"path"
	"strings"

	"github.com/opencontainers/go-digest"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/oci"
	"github.com/open-component-model/ocm/pkg/contexts/oci/artdesc"
	"github.com/open-component-model/ocm/pkg/contexts/oci/repositories/artefactset"
	"github.com/open-component-model/ocm/pkg/contexts/oci/repositories/ocireg"
	"github.com/open-component-model/ocm/pkg/contexts/oci/transfer"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/localblob"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/localociblob"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/ociartefact"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/ociblob"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/compatattr"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/keepblobattr"
	storagecontext "github.com/open-component-model/ocm/pkg/contexts/ocm/blobhandler/oci"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/utils"
)

func init() {
	for _, mime := range artdesc.ArchiveBlobTypes() {
		cpi.RegisterBlobHandler(NewArtefactHandler(OCIRegBaseFunction), cpi.ForRepo(oci.CONTEXT_TYPE, ocireg.Type),
			cpi.ForMimeType(mime))
		cpi.RegisterBlobHandler(NewArtefactHandler(OCIRegBaseFunction), cpi.ForRepo(oci.CONTEXT_TYPE, ocireg.LegacyType),
			cpi.ForMimeType(mime))
		cpi.RegisterBlobHandler(NewArtefactHandler(OCIRegBaseFunction), cpi.ForRepo(oci.CONTEXT_TYPE, ocireg.ShortType),
			cpi.ForMimeType(mime))
	}
	cpi.RegisterBlobHandler(NewBlobHandler(OCIRegBaseFunction), cpi.ForRepo(oci.CONTEXT_TYPE, ocireg.Type))
	cpi.RegisterBlobHandler(NewBlobHandler(OCIRegBaseFunction), cpi.ForRepo(oci.CONTEXT_TYPE, ocireg.LegacyType))
	cpi.RegisterBlobHandler(NewBlobHandler(OCIRegBaseFunction), cpi.ForRepo(oci.CONTEXT_TYPE, ocireg.ShortType))
}

////////////////////////////////////////////////////////////////////////////////

type BaseFunction func(ctx *storagecontext.StorageContext) string

func OCIRegBaseFunction(ctx *storagecontext.StorageContext) string {
	return ctx.Repository.(*ocireg.Repository).GetBaseURL()
}

// blobHandler is the default handling to store local blobs as local blobs but with an additional
// globally accessible OCIBlob access method.
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

func (b *blobHandler) StoreBlob(blob cpi.BlobAccess, artType, hint string, global cpi.AccessSpec, ctx cpi.StorageContext) (cpi.AccessSpec, error) {
	ocictx := ctx.(*storagecontext.StorageContext)

	values := []interface{}{
		"arttype", artType,
		"mediatype", blob.MimeType(),
		"hint", hint,
	}
	if m, ok := blob.(accessio.AnnotatedBlobAccess[cpi.AccessMethod]); ok {
		cpi.BlobHandlerLogger(ctx.GetContext()).Debug("oci blob handler with ocm access source",
			append(values, "sourcetype", m.Source().AccessSpec().GetType())...,
		)
	} else {
		cpi.BlobHandlerLogger(ctx.GetContext()).Debug("oci blob handler", values...)
	}

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

// artefactHandler stores artefact blobs as OCIArtefacts.
type artefactHandler struct {
	blobHandler
}

func NewArtefactHandler(base BaseFunction) cpi.BlobHandler {
	return &artefactHandler{blobHandler{base}}
}

func (b *artefactHandler) StoreBlob(blob cpi.BlobAccess, artType, hint string, global cpi.AccessSpec, ctx cpi.StorageContext) (cpi.AccessSpec, error) {
	mediaType := blob.MimeType()

	if !artdesc.IsOCIMediaType(mediaType) || (!strings.HasSuffix(mediaType, "+tar") && !strings.HasSuffix(mediaType, "+tar+gzip")) {
		return nil, nil
	}

	errhint := "[" + hint + "]"
	log := cpi.BlobHandlerLogger(ctx.GetContext())

	values := []interface{}{
		"arttype", artType,
		"mediatype", mediaType,
		"hint", hint,
	}

	var art oci.ArtefactAccess
	var err error
	var finalizer utils.Finalizer
	defer finalizer.Finalize()

	keep := keepblobattr.Get(ctx.GetContext())

	if m, ok := blob.(accessio.AnnotatedBlobAccess[cpi.AccessMethod]); ok {
		// prepare for optimized point to point implementation
		log.Debug("oci artefact handler with ocm access source",
			append(values, "sourcetype", m.Source().AccessSpec().GetType())...,
		)
		if ocimeth, ok := m.Source().(ociartefact.AccessMethod); !keep && ok {
			art, _, err = ocimeth.GetArtefact(&finalizer)
			if err != nil {
				return nil, errors.Wrapf(err, "cannot access source artefact")
			}
			defer art.Close()
		}
	} else {
		log.Debug("oci artefact handler", values...)
	}

	var namespace oci.NamespaceAccess
	var version string
	var name string
	var tag string
	var digest digest.Digest

	ocictx := ctx.(*storagecontext.StorageContext)
	base := b.GetBaseURL(ocictx)
	if hint == "" {
		namespace = ocictx.Namespace
	} else {
		prefix := cpi.RepositoryPrefix(ctx.TargetComponentRepository().GetSpecification())
		i := strings.LastIndex(hint, ":")
		if i > 0 {
			version = hint[i:]
			tag = version[1:] // remove colon
			name = path.Join(prefix, hint[:i])
		} else {
			name = path.Join(prefix, hint)
		}
		namespace, err = ocictx.Repository.LookupNamespace(name)
		if err != nil {
			return nil, err
		}
		defer namespace.Close()
	}

	errhint += " namespace " + namespace.GetNamespace()

	if art == nil {
		log.Debug("using artefact set transfer mode")
		set, err := artefactset.OpenFromBlob(accessobj.ACC_READONLY, blob)
		if err != nil {
			return nil, wrap(err, errhint, "open blob")
		}
		defer set.Close()
		digest = set.GetMain()
		art, err = set.GetArtefact(digest.String())
		if err != nil {
			return nil, wrap(err, errhint, "get artefact from blob")
		}
		defer art.Close()
	} else {
		log.Debug("using direct transfer mode")
		digest = art.Digest()
	}

	if version == "" {
		version = "@" + digest.String()
	}

	err = transfer.TransferArtefact(art, namespace, oci.AsTags(tag)...)
	if err != nil {
		return nil, wrap(err, errhint, "transfer artefact")
	}

	ref := path.Join(base, namespace.GetNamespace()) + version
	var acc cpi.AccessSpec = ociartefact.New(ref)

	if keep {
		err := ocictx.Manifest.AddBlob(blob)
		if err != nil {
			return nil, wrap(err, errhint, "store local blob")
		}
		err = ocictx.AssureLayer(blob)
		if err != nil {
			return nil, wrap(err, errhint, "assure local blob layer")
		}
		acc = localblob.New(blob.Digest().String(), hint, blob.MimeType(), acc)
	}
	return acc, nil
}

func wrap(err error, msg string, args ...interface{}) error {
	for _, a := range args {
		msg = fmt.Sprintf("%s: %s", msg, a)
	}
	return errors.Wrapf(err, "exploding OCI artefact resource blob (%s)", msg)
}
