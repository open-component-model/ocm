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

package empty

import (
	"github.com/gardener/ocm/pkg/errors"
	"github.com/gardener/ocm/pkg/oci"
)

type Repository struct{}

func NewEmptyRepository() oci.Repository {
	return &Repository{}
}

func (r Repository) ExistsArtefact(name string, version string) (bool, error) {
	return false, nil
}

func (r Repository) LookupArtefact(name string, version string) (oci.ArtefactAccess, error) {
	return nil, oci.ErrUnknownArtefact(name, version)
}

func (r Repository) ComposeArtefact(name string, version string) (oci.ArtefactComposer, error) {
	return nil, errors.ErrNotSupported("artefact composition")
}

func (r Repository) WriteArtefact(access oci.ArtefactAccess) (oci.ArtefactAccess, error) {
	return nil, errors.ErrNotSupported("write access")
}

var _ oci.Repository = &Repository{}
