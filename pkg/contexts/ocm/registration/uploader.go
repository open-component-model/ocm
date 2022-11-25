// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package registration

import (
	"github.com/open-component-model/ocm/pkg/contexts/ocm/internal"
)

type (
	BlobHandlerOption  = internal.BlobHandlerOption
	BlobHandlerConfig  = internal.BlobHandlerConfig
	BlobHandlerOptions = internal.BlobHandlerOptions
)

func RegisterBlobHandlerByName(ctx internal.Context, name string, config BlobHandlerConfig, opts ...BlobHandlerOption) error {
	_, err := ctx.BlobHandlers().RegisterByName(name, ctx, config, opts...)
	return err
}

func WithPrio(prio int) BlobHandlerOption {
	return internal.WithPrio(prio)
}

func ForArtifactType(t string) BlobHandlerOption {
	return internal.ForArtifactType(t)
}

func ForMimeType(t string) BlobHandlerOption {
	return internal.ForMimeType(t)
}
