// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package registration

import (
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/download"
	"github.com/open-component-model/ocm/pkg/errors"
)

func RegisterDownloadHandler(ctx cpi.Context, hdlr download.Handler, olist ...cpi.BlobHandlerOption) error {
	opts := cpi.NewBlobHandlerOptions(olist...)
	if opts.Priority > 0 {
		return errors.ErrInvalid("option", "priority")
	}
	download.For(ctx).Register(opts.ArtifactType, opts.MimeType, hdlr)
	return nil
}
