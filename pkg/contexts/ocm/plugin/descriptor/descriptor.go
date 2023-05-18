// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package descriptor

const VERSION = "v1"

type Descriptor struct {
	Version       string `json:"version,omitempty"`
	PluginName    string `json:"pluginName"`
	PluginVersion string `json:"pluginVersion"`
	Short         string `json:"shortDescription"`
	Long          string `json:"description"`

	Actions       []ActionDescriptor         `json:"actions,omitempty"`
	AccessMethods []AccessMethodDescriptor   `json:"accessMethods,omitempty"`
	Uploaders     List[UploaderDescriptor]   `json:"uploaders,omitempty"`
	Downloaders   List[DownloaderDescriptor] `json:"downloaders,omitempty"`
}

type DownloaderKey = ArtifactContext

func NewDownloaderKey(arttype, mediatype string) DownloaderKey {
	return DownloaderKey{
		ArtifactType: arttype,
		MediaType:    mediatype,
	}
}

type DownloaderDescriptor struct {
	Name             string          `json:"name"`
	Description      string          `json:"description"`
	Constraints      []DownloaderKey `json:"constraints,omitempty"`
	ConfigScheme     string          `json:"configScheme,omitempty"`
	AutoRegistration []DownloaderKey `json:"autoRegistration,omitempty"`
}

func (d DownloaderDescriptor) GetName() string {
	return d.Name
}

func (d DownloaderDescriptor) GetDescription() string {
	return d.Description
}

func (d DownloaderDescriptor) GetConstraints() []DownloaderKey {
	return d.Constraints
}

type UploaderDescriptor struct {
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Constraints []UploaderKey `json:"constraints,omitempty"`
}

func (d UploaderDescriptor) GetName() string {
	return d.Name
}

func (d UploaderDescriptor) GetDescription() string {
	return d.Description
}

func (d UploaderDescriptor) GetConstraints() []UploaderKey {
	return d.Constraints
}

type AccessMethodDescriptor struct {
	Name        string      `json:"name"`
	Version     string      `json:"version,omitempty"`
	Description string      `json:"description"`
	Format      string      `json:"format"`
	CLIOptions  []CLIOption `json:"options,omitempty"`
}

type ActionDescriptor struct {
	Name             string   `json:"name"`
	Versions         []string `json:"versions,omitempty"`
	Description      string   `json:"description,omitempty"`
	ConsumerType     string   `json:"consumerType,omitempty"`
	DefaultSelectors []string `json:"defaultSelectors,omitempty"`
}

type CLIOption struct {
	Name        string `json:"name"`
	Type        string `json:"type,omitempty"`
	Description string `json:"description,omitempty"`
}
