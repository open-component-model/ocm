// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package dirtree

import (
	"fmt"

	"github.com/mandelsoft/vfs/pkg/layerfs"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"

	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/common/compression"
	"github.com/open-component-model/ocm/pkg/contexts/oci/artdesc"
	"github.com/open-component-model/ocm/pkg/contexts/oci/repositories/artifactset"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/download"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/resourcetypes"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/finalizer"
	"github.com/open-component-model/ocm/pkg/generics"
	"github.com/open-component-model/ocm/pkg/utils"
)

type Handler struct {
	configtypes generics.Set[string]
	archive     bool
}

func New(types ...string) *Handler {
	return &Handler{
		configtypes: generics.NewSet[string](types...),
	}
}

func NewAsArchive(types ...string) *Handler {
	return &Handler{
		configtypes: generics.NewSet[string](types...),
		archive:     true,
	}
}

var DefaultHandler = New(artdesc.MediaTypeImageConfig)

func init() {
	download.Register(resourcetypes.DIRECTORY_TREE, artifactset.MediaType(artdesc.MediaTypeImageManifest), DefaultHandler)
	download.Register(resourcetypes.FILESYSTEM_LEGACY, artifactset.MediaType(artdesc.MediaTypeImageManifest), DefaultHandler)
}

func (h *Handler) Download(p common.Printer, racc cpi.ResourceAccess, path string, fs vfs.FileSystem) (ok bool, dest string, err error) {
	var finalize finalizer.Finalizer

	defer finalize.FinalizeWithErrorPropagation(&err)
	lfs, err := h.GetFilesystemForResource(racc)
	if err != nil || lfs == nil {
		return err != nil, "", err
	}
	finalize.With(func() error { return vfs.Cleanup(lfs) })
	if h.archive {
		w, err := fs.OpenFile(path, vfs.O_TRUNC|vfs.O_CREATE|vfs.O_WRONLY, 0o600)
		if err != nil {
			return true, "", errors.Wrapf(err, "cannot write target archive %s", path)
		}
		finalize.Close(w)
		return true, path, utils.PackFsIntoTar(lfs, w)
	} else {
		err := fs.MkdirAll(path, 0o700)
		if err != nil {
			return true, "", errors.Wrapf(err, "cannot create target directory")
		}
		return true, path, vfs.CopyDir(lfs, "/", fs, path)
	}
}

// GetFilesystemForResource provides a virtual filesystem for an OCi image manifest
// provided by the given resource matching the configured config types.
// It returns nil without error, if the OCI artifact does not match the requirement.
func (h *Handler) GetFilesystemForResource(racc cpi.ResourceAccess) (fs vfs.FileSystem, err error) {
	var finalize finalizer.Finalizer
	defer finalize.FinalizeWithErrorPropagation(&err)

	r, err := ocm.ResourceReader(racc)
	if err != nil {
		return nil, err
	}
	finalize.WithVoid(func() { r.Close() }) // TODO: close handling for ReaderOption
	set, err := artifactset.Open(accessobj.ACC_READONLY, "", 0, accessio.Reader(r))
	if err != nil {
		return nil, err
	}
	finalize.Close(set)
	return h.GetFilesystemForArtifactSet(set)
}

// GetFilesystemForArtifactSet provides a virtual filesystem for an OCi image manifest
// provided by the given artifact set matching the configured config types.
// It returns nil without error, if the OCI artifact does not match the requirement.
func (h *Handler) GetFilesystemForArtifactSet(set *artifactset.ArtifactSet) (fs vfs.FileSystem, err error) {
	var finalize finalizer.Finalizer
	defer finalize.FinalizeWithErrorPropagation(&err)

	m, err := set.GetArtifact(set.GetMain().String())
	if !m.IsManifest() {
		return nil, fmt.Errorf("oci artifact is no image manifest")
	}
	finalize.Close(m)
	macc := m.ManifestAccess()
	if !h.configtypes.Contains(macc.GetDescriptor().Config.MediaType) {
		return nil, nil
	}

	var cfs vfs.FileSystem
	finalize.With(func() error {
		return vfs.Cleanup(cfs)
	})

	// setup layered filesystem from manifest layers
	for i, l := range macc.GetDescriptor().Layers {
		nested := finalize.Nested()

		blob, err := macc.GetBlob(l.Digest)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot get blob for layer %d", i)
		}
		nested.Close(blob)
		r, err := blob.Reader()
		if err != nil {
			return nil, errors.Wrapf(err, "cannot get reader for layer blob %d", i)
		}
		nested.Close(r)
		r, _, err = compression.AutoDecompress(r)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot determine compression for layer blob %d", i)
		}
		nested.Close(r)

		fslayer, err := osfs.NewTempFileSystem()
		if err != nil {
			return nil, errors.Wrapf(err, "cannot create filesystem for layer %d", i)
		}
		nested.With(func() error {
			return vfs.Cleanup(fslayer)
		})
		err = utils.ExtractTarToFs(fslayer, r)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot unpack layer blob %d", i)
		}

		if cfs == nil {
			cfs = fslayer
		} else {
			cfs = layerfs.New(fslayer, cfs)
		}
		fslayer = nil // don't cleanup used layer
		nested.Finalize()
	}
	fs = cfs
	cfs = nil // don't cleanup used filesystem
	return fs, nil
}
