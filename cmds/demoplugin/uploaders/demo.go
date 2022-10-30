// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package uploaders

import (
	"io"
	"os"
	"path/filepath"

	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/ppi"

	"github.com/open-component-model/ocm/cmds/demoplugin/accessmethods"
)

type Uploader struct {
	ppi.UploaderBase
}

var _ ppi.Uploader = (*Uploader)(nil)

func New() ppi.Uploader {
	return &Uploader{
		UploaderBase: ppi.MustNewUploaderBase("demo", "upload temp files"),
	}
}

func (a *Uploader) Writer(p ppi.Plugin, arttype, mediatype, hint string, creds credentials.Credentials) (io.WriteCloser, ppi.AccessSpecProvider, error) {
	var file *os.File
	var err error
	if hint == "" {
		file, err = os.CreateTemp(os.TempDir(), "demo.*.blob")
	} else {
		file, err = os.OpenFile(filepath.Join(os.TempDir(), hint), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
	}
	if err != nil {
		return nil, nil, err
	}
	writer := NewWriter(file, mediatype, hint == "", accessmethods.NAME, accessmethods.VERSION)
	return writer, writer.Specification, nil
}
