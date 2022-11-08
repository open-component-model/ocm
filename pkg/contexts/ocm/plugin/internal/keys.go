// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"fmt"
)

type RepositoryContext struct {
	ContextType    string `json:"contextType"`
	RepositoryType string `json:"repositoryType"`
}

func (k RepositoryContext) HasRepo() bool {
	return k.ContextType != "" || k.RepositoryType != ""
}

func (k RepositoryContext) IsValid() bool {
	return k.HasRepo() || (k.ContextType == "" && k.RepositoryType == "")
}

func (k RepositoryContext) String() string {
	if k.HasRepo() {
		return fmt.Sprintf("[%s:%s]", k.ContextType, k.RepositoryType)
	}
	return ""
}

type ArtefactContext struct {
	ArtifactType string `json:"artifactType"`
	MediaType    string `json:"mediaType"`
}

func (k ArtefactContext) IsValid() bool {
	return k.ArtifactType != "" || k.MediaType != ""
}

func (k ArtefactContext) GetArtefactType() string {
	return k.ArtifactType
}

func (k ArtefactContext) GetMediaType() string {
	return k.MediaType
}

func (k ArtefactContext) String() string {
	return fmt.Sprintf("%s:%s", k.ArtifactType, k.MediaType)
}

func (k ArtefactContext) SetArtefact(arttype, mediatype string) ArtefactContext {
	k.ArtifactType = arttype
	k.MediaType = mediatype
	return k
}

type UploaderKey struct {
	RepositoryContext `json:",inline"`
	ArtefactContext   `json:",inline"`
}

func (k UploaderKey) IsValid() bool {
	return k.ArtefactContext.IsValid() && k.RepositoryContext.IsValid()
}

func (k UploaderKey) String() string {
	return fmt.Sprintf("%s%s", k.ArtefactContext.String(), k.RepositoryContext.String())
}

func (k UploaderKey) SetArtefact(arttype, mediatype string) UploaderKey {
	k.ArtifactType = arttype
	k.MediaType = mediatype
	return k
}

func (k UploaderKey) SetRepo(contexttype, repotype string) UploaderKey {
	k.ContextType = contexttype
	k.RepositoryType = repotype
	return k
}
