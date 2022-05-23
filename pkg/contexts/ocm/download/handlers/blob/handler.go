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

	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/download"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/out"
)

type Handler struct{}

func init() {
	download.Register(download.ALL, &Handler{})
}

func wrapErr(err error, racc cpi.ResourceAccess) error {
	if err == nil {
		return nil
	}
	m := racc.Meta()
	return errors.Wrapf(err, "resource %s/%s%s", m.GetName(), m.GetVersion(), m.ExtraIdentity.String())
}

func (_ Handler) Download(ctx out.Context, racc cpi.ResourceAccess, path string, fs vfs.FileSystem) (bool, error) {
	rd, err := cpi.ResourceReader(racc)
	if err != nil {
		return true, wrapErr(err, racc)
	}
	defer rd.Close()
	file, err := fs.OpenFile(path, vfs.O_TRUNC|vfs.O_CREATE|vfs.O_WRONLY, 0660)
	if err != nil {
		return true, wrapErr(errors.Wrapf(err, "creating target file %q", path), racc)
	}
	defer file.Close()
	n, err := io.Copy(file, rd)
	if err == nil {
		out.Outf(ctx, "%s: %d byte(s) written\n", path, n)
	}
	return true, wrapErr(err, racc)
}
