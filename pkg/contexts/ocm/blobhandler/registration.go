// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package blobhandler

import (
	"fmt"

	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
)

type (
	BlobHandlerOption   = cpi.BlobHandlerOption
	BlobHandlerConfig   = cpi.BlobHandlerConfig
	BlobHandlerOptions  = cpi.BlobHandlerOptions
	BlobHandlerRegistry = cpi.BlobHandlerRegistry
)

func For(ctx cpi.ContextProvider) cpi.BlobHandlerRegistry {
	return ctx.OCMContext().BlobHandlers()
}

func RegisterHandlerByName(ctx cpi.ContextProvider, name string, config BlobHandlerConfig, opts ...BlobHandlerOption) error {
	o, err := For(ctx).RegisterByName(name, ctx.OCMContext(), config, opts...)
	if err != nil {
		return err
	}
	if !o {
		return fmt.Errorf("no matching handler found for %q", name)
	}
	return nil
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
