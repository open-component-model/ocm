// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package registration

import (
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
)

func RegisterBlobHandlerByName(ctx cpi.Context, name string, config cpi.BlobHandlerConfig, opts ...cpi.BlobHandlerOption) error {
	hdlrs := ctx.BlobHandlers()
	_, err := hdlrs.RegisterByName(name, ctx, config, opts...)
	return err
}

func WithPrio(prio int) cpi.BlobHandlerOption {
	return cpi.WithPrio(prio)
}

func ForArtifactType(t string) cpi.BlobHandlerOption {
	return cpi.ForArtifactType(t)
}

func ForMimeType(t string) cpi.BlobHandlerOption {
	return cpi.ForMimeType(t)
}
