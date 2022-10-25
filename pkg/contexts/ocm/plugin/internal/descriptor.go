// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package internal

const VERSION = "v1"

type Descriptor struct {
	Version       string `json:"version,omitempty"`
	PluginName    string `json:"pluginName"`
	PluginVersion string `json:"pluginVersion"`
	Short         string `json:"shortDescription"`
	Long          string `json:"description"`

	AccessMethods []AccessMethodDescriptor `json:"accessMethods,omitempty"`
}

type AccessMethodDescriptor struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
	Short   string `json:"shortDescription"`
	Long    string `json:"description"`
}
