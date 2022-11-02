// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"fmt"
)

const VERSION = "v1"

type Descriptor struct {
	Version       string `json:"version,omitempty"`
	PluginName    string `json:"pluginName"`
	PluginVersion string `json:"pluginVersion"`
	Short         string `json:"shortDescription"`
	Long          string `json:"description"`

	AccessMethods []AccessMethodDescriptor `json:"accessMethods,omitempty"`
	Uploaders     []UploaderDescriptor     `json:"uploaders,omitempty"`
}

type UploaderKey struct {
	ContextType    string `json:"contextType"`
	RepositoryType string `json:"repositoryType"`
	ArtifactType   string `json:"artifactType"`
	MediaType      string `json:"mediaType"`
}

func (k UploaderKey) String() string {
	if k.RepositoryType != "" || k.ContextType != "" {
		return fmt.Sprintf("%s:%s[%s:%s]", k.ContextType, k.RepositoryType, k.ArtifactType, k.MediaType)
	}
	return fmt.Sprintf("%s:%s", k.ArtifactType, k.MediaType)
}

type UploaderDescriptor struct {
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Costraints  []UploaderKey `json:"constraints"`
}

type AccessMethodDescriptor struct {
	Name        string `json:"name"`
	Version     string `json:"version,omitempty"`
	Description string `json:"description"`
	Format      string `json:"format"`
}
