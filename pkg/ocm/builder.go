// Copyright 2022 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package ocm

import (
	"context"

	"github.com/open-component-model/ocm/pkg/credentials"
	"github.com/open-component-model/ocm/pkg/datacontext"
	"github.com/open-component-model/ocm/pkg/oci"
	"github.com/open-component-model/ocm/pkg/ocm/core"
)

func WithContext(ctx context.Context) core.Builder {
	return core.Builder{}.WithContext(ctx)
}

func WithSharedAttributes(ctx datacontext.AttributesContext) core.Builder {
	return core.Builder{}.WithSharedAttributes(ctx)
}

func WithCredentials(ctx credentials.Context) core.Builder {
	return core.Builder{}.WithCredentials(ctx)
}

func WithOCIRepositories(ctx oci.Context) core.Builder {
	return core.Builder{}.WithOCIRepositories(ctx)
}

func WithRepositoyTypeScheme(scheme RepositoryTypeScheme) core.Builder {
	return core.Builder{}.WithRepositoyTypeScheme(scheme)
}

func WithAccessypeScheme(scheme AccessTypeScheme) core.Builder {
	return core.Builder{}.WithAccessTypeScheme(scheme)
}

func WithRepositorySpecHandlers(reg RepositorySpecHandlers) core.Builder {
	return core.Builder{}.WithRepositorySpecHandlers(reg)
}

func WithBlobHandlers(reg BlobHandlerRegistry) core.Builder {
	return core.Builder{}.WithBlobHandlers(reg)
}

func WithBlobDigesters(reg BlobDigesterRegistry) core.Builder {
	return core.Builder{}.WithBlobDigesters(reg)
}

func New() Context {
	return core.Builder{}.New()
}
