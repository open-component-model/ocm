// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package registry

import (
	"fmt"
)

type RegistrationKey struct {
	ArtefactType string `json:"artefactType"`
	MediaType    string `json:"mediaType"`
}

// check type match for constraint type.
var _ Registry[string, RegistrationKey]

func (k RegistrationKey) IsValid() bool {
	return k.ArtefactType != "" || k.MediaType != ""
}

func (k RegistrationKey) GetArtefactType() string {
	return k.ArtefactType
}

func (k RegistrationKey) GetMediaType() string {
	return k.MediaType
}

func (k RegistrationKey) SetArtefact(arttype, medtatype string) RegistrationKey {
	k.ArtefactType = arttype
	k.MediaType = medtatype
	return k
}

func (k RegistrationKey) Key() RegistrationKey {
	return k
}

func (k RegistrationKey) String() string {
	return fmt.Sprintf("%s:%s", k.ArtefactType, k.MediaType)
}
