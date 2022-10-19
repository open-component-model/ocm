// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"

	"github.com/open-component-model/ocm/pkg/contexts/config/core"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext"
)

func WithContext(ctx context.Context) core.Builder {
	return core.Builder{}.WithContext(ctx)
}

func WithSharedAttributes(ctx datacontext.AttributesContext) core.Builder {
	return core.Builder{}.WithSharedAttributes(ctx)
}

func WithConfigTypeScheme(scheme ConfigTypeScheme) core.Builder {
	return core.Builder{}.WithConfigTypeScheme(scheme)
}

func New(mode ...datacontext.BuilderMode) Context {
	return core.Builder{}.New(mode...)
}
