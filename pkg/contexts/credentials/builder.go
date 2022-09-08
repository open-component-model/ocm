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

package credentials

import (
	"context"

	"github.com/open-component-model/ocm/pkg/contexts/config"
	"github.com/open-component-model/ocm/pkg/contexts/credentials/core"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext"
)

func WithContext(ctx context.Context) core.Builder {
	return core.Builder{}.WithContext(ctx)
}

func WithConfigs(ctx config.Context) core.Builder {
	return core.Builder{}.WithConfig(ctx)
}

func WithRepositoyTypeScheme(scheme RepositoryTypeScheme) core.Builder {
	return core.Builder{}.WithRepositoyTypeScheme(scheme)
}

func WithStandardConumerMatchers(matchers core.IdentityMatcherRegistry) core.Builder {
	return core.Builder{}.WithStandardConumerMatchers(matchers)
}

func New(mode ...datacontext.BuilderMode) Context {
	return core.Builder{}.New(mode...)
}
