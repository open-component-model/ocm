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

package v2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	compdesc "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/versions/v2"
)

type ComponentDescriptorStatus struct{}

// +kubebuilder:object:root=true
// +kubebuilder:storageversion

type ComponentDescriptor struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   compdesc.ComponentSpec    `json:"spec,omitempty"`
	Status ComponentDescriptorStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ComponentDescriptorList contains a list of Component Descriptors.
type ComponentDescriptorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ComponentDescriptor `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ComponentDescriptor{}, &ComponentDescriptorList{})
}
