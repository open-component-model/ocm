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

package ocm_gardener_cloud

import (
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
)

const GROUP = "ocm.gardener.cloud"
const KIND = "ComponentVersion"

// TypeMeta describes the schema of a descriptor
type TypeMeta struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
}

// ObjectMeta defines the metadata of the component descriptor.
type ObjectMeta struct {
	// Name is the name of the component.
	Name string `json:"name"`
	// Version is the version of the component.
	Version string `json:"version"`
	// Labels describe additional properties of the component version
	Labels metav1.Labels `json:"labels,omitempty"`
	// Provider described the component provider
	Provider Provider `json:"provider"`
}

// GetName returns the name of the object.
func (o ObjectMeta) GetName() string {
	return o.Name
}

// SetName sets the name of the object.
func (o *ObjectMeta) SetName(name string) {
	o.Name = name
}

// GetVersion returns the version of the object.
func (o ObjectMeta) GetVersion() string {
	return o.Version
}

// SetVersion sets the version of the object.
func (o *ObjectMeta) SetVersion(version string) {
	o.Version = version
}

// GetLabels returns the label of the object.
func (o ObjectMeta) GetLabels() metav1.Labels {
	return o.Labels
}

// SetLabels sets the labels of the object.
func (o *ObjectMeta) SetLabels(labels []metav1.Label) {
	o.Labels = labels
}

// Provider describes the provider information of a component version
type Provider struct {
	Name string `json:"name"`
	// Labels describe additional properties of provider
	Labels metav1.Labels `json:"labels,omitempty"`
}

// GetName returns the name of the provider.
func (o Provider) GetName() string {
	return o.Name
}

// SetName sets the name of the provider.
func (o *Provider) SetName(name string) {
	o.Name = name
}

// GetLabels returns the label of the provider.
func (o Provider) GetLabels() metav1.Labels {
	return o.Labels
}

// SetLabels sets the labels of the provider.
func (o *Provider) SetLabels(labels []metav1.Label) {
	o.Labels = labels
}
