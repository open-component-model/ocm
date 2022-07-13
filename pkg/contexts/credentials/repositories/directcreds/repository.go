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

package directcreds

import (
	"github.com/open-component-model/ocm/pkg/contexts/credentials/cpi"
	"github.com/open-component-model/ocm/pkg/errors"
)

type Repository struct {
	Credentials cpi.Credentials
}

func NewRepository(creds cpi.Credentials) cpi.Repository {
	return &Repository{
		Credentials: creds,
	}
}

func (r *Repository) ExistsCredentials(name string) (bool, error) {
	return name == DirectCredentialsRepositoryType, nil
}

func (r *Repository) LookupCredentials(name string) (cpi.Credentials, error) {
	if name != DirectCredentialsRepositoryType && name != "" {
		return nil, cpi.ErrUnknownCredentials(name)
	}
	return r.Credentials, nil
}

func (r *Repository) WriteCredentials(name string, creds cpi.Credentials) (cpi.Credentials, error) {
	return nil, errors.ErrNotSupported(cpi.KIND_CREDENTIALS, "write", "constant credential")
}

var _ cpi.Repository = &Repository{}
