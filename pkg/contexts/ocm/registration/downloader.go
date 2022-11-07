// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package registration

import (
	ocmcpi "github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/download"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/internal"
	"github.com/open-component-model/ocm/pkg/errors"
)

func RegisterDownloadHandler(ctx internal.Context, hdlr download.Handler, olist ...BlobHandlerOption) error {
	opts := ocmcpi.NewBlobHandlerOptions(olist...)
	if opts.Priority > 0 {
		return errors.ErrInvalid("option", "priority")
	}
	download.For(ctx).Register(opts.ArtefactType, opts.MimeType, hdlr)
	return nil
}
