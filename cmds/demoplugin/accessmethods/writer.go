// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package accessmethods

import (
	"os"
	"path/filepath"

	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/ppi"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/runtime"
)

type writer = accessio.DigestWriter

type Writer struct {
	*writer
	file    *os.File
	name    string
	version string
	media   string
	spec    *AccessSpec
}

func NewWriter(file *os.File, media, name, version string) *Writer {
	return &Writer{
		writer:  accessio.NewDefaultDigestWriter(file),
		file:    file,
		name:    name,
		version: version,
		media:   media,
	}
}

func (w *Writer) Close() error {
	err := w.writer.Close()
	if err == nil {
		n := filepath.Join(os.TempDir(), common.DigestToFileName(w.writer.Digest()))
		err := os.Rename(w.file.Name(), n)
		if err != nil {
			return errors.Wrapf(err, "cannot rename %q to %q", w.file.Name(), n)
		}
		w.spec = &AccessSpec{
			ObjectVersionedType: runtime.NewVersionedObjectType(w.name, w.version),
			Path:                n,
			MediaType:           w.media,
		}
	}
	return err
}

func (w *Writer) Specification() ppi.AccessSpec {
	return w.spec
}
