// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package plugin

import (
	"github.com/opencontainers/go-digest"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/ppi/cmds/accessmethod"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/ppi/cmds/accessmethod/get"
)

type AccessDataWriter struct {
	plugin  Plugin
	acctype string
}

func NewAccessDataWriter(p Plugin, acctype string) *AccessDataWriter {
	return &AccessDataWriter{p, acctype}
}

func (d *AccessDataWriter) WriteTo(w accessio.Writer) (int64, digest.Digest, error) {
	dw := accessio.NewDefaultDigestWriter(accessio.NopWriteCloser(w))
	_, err := d.plugin.Exec(nil, dw, accessmethod.NAME, get.Name, d.acctype)
	if err != nil {
		return accessio.BLOB_UNKNOWN_SIZE, accessio.BLOB_UNKNOWN_DIGEST, err
	}
	return dw.Size(), dw.Digest(), nil
}
