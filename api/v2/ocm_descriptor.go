// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

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
