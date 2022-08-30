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

package blob

import (
	"io"
	"strings"

	"github.com/mandelsoft/vfs/pkg/vfs"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/oci/artdesc"
	"github.com/open-component-model/ocm/pkg/contexts/oci/repositories/artefactset"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/consts"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/download"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/mime"
	"github.com/open-component-model/ocm/pkg/out"
)

const TYPE = consts.HelmChart

type Handler struct{}

func init() {
	download.Register(TYPE, &Handler{})
}

func (h Handler) Download(ctx out.Context, racc cpi.ResourceAccess, path string, fs vfs.FileSystem) (bool, string, error) {
	meth, err := racc.AccessMethod()
	if err != nil {
		return false, "", err
	}
	defer meth.Close()
	if mime.BaseType(meth.MimeType()) != mime.BaseType(artdesc.MediaTypeImageManifest) {
		return false, "", nil
	}
	rd, err := meth.Reader()
	if err != nil {
		return true, "", err
	}
	defer rd.Close()
	set, err := artefactset.Open(accessobj.ACC_READONLY, "", 0, accessio.Reader(rd))
	if err != nil {
		return true, "", err
	}
	art, err := set.GetArtefact(set.GetMain().String())
	if err != nil {
		return true, "", err
	}
	m := art.ManifestAccess()
	if m == nil {
		return true, "", errors.Newf("artefact is no image manifest")
	}
	if len(m.GetDescriptor().Layers) < 1 {
		return true, "", errors.Newf("no layers found")
	}
	if !strings.HasSuffix(path, ".tgz") {
		path = path + ".tgz"
	}
	blob, err := m.GetBlob(m.GetDescriptor().Layers[0].Digest)
	if err != nil {
		return true, "", err
	}
	err = h.write(ctx, blob, path, fs)
	if err != nil {
		return true, "", err
	}
	if len(m.GetDescriptor().Layers) > 1 {
		path = path[:len(path)-3] + "prov"
		blob, err := m.GetBlob(m.GetDescriptor().Layers[1].Digest)
		if err != nil {
			return true, "", err
		}
		err = h.write(ctx, blob, path, fs)
		if err != nil {
			return true, "", err
		}
	}
	return true, path, nil
}

func (_ Handler) write(ctx out.Context, blob accessio.BlobAccess, path string, fs vfs.FileSystem) error {
	cr, err := blob.Reader()
	if err != nil {
		return err
	}
	defer cr.Close()
	file, err := fs.OpenFile(path, vfs.O_TRUNC|vfs.O_CREATE|vfs.O_WRONLY, 0o660)
	if err != nil {
		return err
	}
	defer file.Close()
	n, err := io.Copy(file, cr)
	if err == nil {
		out.Outf(ctx, "%s: %d byte(s) written\n", path, n)
	}
	return nil
}
