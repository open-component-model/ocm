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

package core

import (
	"github.com/modern-go/reflect2"
)

// CredentialsSource is a factory for effective credentials.
type CredentialsSource interface {
	Credentials(Context, ...CredentialsSource) (Credentials, error)
}

// CredentialsChain is a chain of credentials, where the
// credential i+1 (is present) is used to resolve credential i.
type CredentialsChain []CredentialsSource

var _ CredentialsSource = CredentialsChain{}

func (c CredentialsChain) Credentials(ctx Context, creds ...CredentialsSource) (Credentials, error) {
	if len(c) == 0 || reflect2.IsNil(c[0]) {
		return nil, nil
	}

	if len(creds) == 0 {
		return c[0].Credentials(ctx, c[1:]...)
	}
	return c[0].Credentials(ctx, append(append(c[:0:len(c)-1+len(creds)], c[1:]...), creds...))
}
