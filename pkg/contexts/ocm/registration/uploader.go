// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package registration

import (
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
)

type (
	BlobHandlerOption  = cpi.BlobHandlerOption
	BlobHandlerConfig  = cpi.BlobHandlerConfig
	BlobHandlerOptions = cpi.BlobHandlerOptions
)

func RegisterBlobHandlerByName(ctx cpi.Context, name string, config BlobHandlerConfig, opts ...BlobHandlerOption) error {
	hdlrs := ctx.BlobHandlers()
	_, err := hdlrs.RegisterByName(name, ctx, config, opts...)
	return err
}

func WithPrio(prio int) BlobHandlerOption {
	return cpi.WithPrio(prio)
}

func ForArtifactType(t string) BlobHandlerOption {
	return cpi.ForArtifactType(t)
}

func ForMimeType(t string) BlobHandlerOption {
	return cpi.ForMimeType(t)
}
