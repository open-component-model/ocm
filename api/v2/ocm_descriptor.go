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
)

// TODO: Not sure what to make with this.
type Object struct {
	// +kubebuilder:validation:Required
	Meta Meta `json:"meta"`
	// +kubebuilder:validation:Required
	Component Component `json:"component"`

	Signatures []Signature `json:"signatures,omitempty"`
}

type Meta struct {
	// +kubebuilder:validation:Required
	SchemaVersion string `json:"schemaVersion"`
}

type Label struct {
	Name   string `json:"name,omitempty"`
	Values string `json:"values,omitempty"`
}

type IdentityAttribute struct {
	IdentityAttributeKeys []IdentityAttributeKey `json:"identityAttributeKeys,omitempty"`
}

type RepositoryContext struct {
	// +kubebuilder:validation:Required
	Type string `json:"type"`
}

type Access struct {
	// +kubebuilder:validation:Required
	// Type is the base type for the access of a source or resource
	Type string `json:"type"`
}

type DigestSpec struct {
	// +kubebuilder:validation:Required
	HashAlgorithm string `json:"hashAlgorithm"`
	// +kubebuilder:validation:Required
	NormalisationAlgorithm string `json:"normalisationAlgorithm"`
	// +kubebuilder:validation:Required
	Value string `json:"value"`
}

type SignatureSpec struct {
	// +kubebuilder:validation:Required
	Algorithm string `json:"algorithm"`
	// +kubebuilder:validation:Required
	Value string `json:"value"`
	// MediaType is the media type of the signature value
	// +kubebuilder:validation:Required
	MediaType string `json:"mediaType"`
}

type Signature struct {
	// +kubebuilder:validation:Required
	Name string `json:"name"`
	// +kubebuilder:validation:Required
	Digest DigestSpec `json:"digest"`
	// +kubebuilder:validation:Required
	Signature SignatureSpec `json:"signature"`
}

// LocalBlobAccess identifier of the local blob within the current component descriptor.
type LocalBlobAccess struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum:=localBlob
	Type string `json:"type"`
	// +kubebuilder:validation:Required
	Digest string `json:"digest"`
}

type Source struct {
	// +kubebuilder:validation:Required
	Name IdentityAttributeKey `json:"name"`
	// +kubebuilder:validation:Required
	Version RelaxedSemver `json:"version"`
	// +kubebuilder:validation:Required
	Type string `json:"type"`
	// TODO: Access can be either of type Access or LocalBlobAccess. We can do this by making this
	// field use *apiextensions.JSON which validates any valid JSON. Then, we can unmarshall it and
	// use the values and not loose type safety. Or, have an explicit field called LocalBlobAccess
	// which can either be provided or not.
	// +kubebuilder:validation:Required
	Access          Access          `json:"access,omitempty"`
	LocalBlobAccess LocalBlobAccess `json:"localBlobAccess,omitempty"`

	ExtraIdentity IdentityAttribute `json:"extraIdentity,omitempty"`
	Labels        []Label           `json:"labels,omitempty"`
}

// RelaxedSemver taken from semver.org and adjusted to allow an optional leading 'v', major-only, and major.minor-only
// this means the following strings are all valid relaxedSemvers:
// 1.2.3
// 1.2.3-foo+bar
// v1.2.3
// v1.2.3-foo+bar
// 1.2
// 1
// v1
// v1.2
// v1-foo+bar
// +kubebuilder:validation:Pattern:=`^[v]?(0|[1-9]\d*)(?:\.(0|[1-9]\d*))?(?:\.(0|[1-9]\d*))?(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`
type RelaxedSemver string

// ComponentReference is a reference to a component.
type ComponentReference struct {
	// +kubebuilder:validation:Required
	Name IdentityAttributeKey `json:"name"`
	// +kubebuilder:validation:Required
	ComponentName string `json:"componentName"`
	// +kubebuilder:validation:Required
	Version RelaxedSemver `json:"version"`

	ExtraIdentity IdentityAttribute `json:"extraIdentity,omitempty"`
	Labels        []Label           `json:"labels,omitempty"`
	Digest        DigestSpec        `json:"digest,omitempty"`
}

// SourceReferences a reference to a (component-local) source
type SourceReferences struct {
	Name          IdentityAttributeKey `json:"name,omitempty"`
	ExtraIdentity IdentityAttribute    `json:"extraIdentity,omitempty"`
}

// Resource is the base type for resources.
type Resource struct {
	// +kubebuilder:validation:Required
	Name IdentityAttributeKey `json:"name"`
	// +kubebuilder:validation:Required
	Version RelaxedSemver `json:"version"`
	// +kubebuilder:validation:Required
	Type string `json:"type"`
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum:=local;external
	Relation string `json:"relation,omitempty"`

	// TODO: Access can be either of type Access or LocalBlobAccess. We can do this by making this
	// field use *apiextensions.JSON which validates any valid JSON. Then, we can unmarshall it and
	// use the values and not loose type safety. Or, have an explicit field called LocalBlobAccess
	// which can either be provided or not.
	// +kubebuilder:validation:Required
	Access          Access          `json:"access,omitempty"`
	LocalBlobAccess LocalBlobAccess `json:"localBlobAccess,omitempty"`

	SrcRefs       []SourceReferences `json:"srcRefs,omitempty"`
	ExtraIdentity IdentityAttribute  `json:"extraIdentity,omitempty"`
	Labels        []Label            `json:"labels,omitempty"`
	Digest        DigestSpec         `json:"digest,omitempty"`
}

// ComponentName (s) MUST start with a valid domain name (as specified by RFC-1034, RFC-1035) with an optional URL path suffix (as specified by RFC-1738)'
// +kubebuilder:validation:MaxLength=255
// +kubebuilder:validation:Pattern:=`^[a-z][-a-z0-9]*([.][a-z][-a-z0-9]*)*[.][a-z]{2,}(/[a-z][-a-z0-9_]*([.][a-z][-a-z0-9_]*)*)+$`
type ComponentName string

// Component is a component.
type Component struct {
	// +kubebuilder:validation:Required
	Name ComponentName `json:"name"`
	// +kubebuilder:validation:Required
	Version RelaxedSemver `json:"version"`
	// +kubebuilder:validation:Required
	RepositoryContexts []RepositoryContext `json:"repositoryContexts,omitempty"`
	// +kubebuilder:validation:Required
	Provider string `json:"provider,omitempty"`
	// +kubebuilder:validation:Required
	Sources []Source `json:"sources,omitempty"`
	// +kubebuilder:validation:Required
	ComponentReferences []ComponentReference `json:"componentReferences,omitempty"`
	// +kubebuilder:validation:Required
	Resources []Resource `json:"resources,omitempty"`

	Labels []Label `json:"labels,omitempty"`
}

type ComponentDescriptorSpec struct {
	Meta  Meta    `json:"meta,omitempty"`
	Label []Label `json:"label,omitempty"`
	// ComponentName MUST start with a valid domain name (as specified by RFC-1034, RFC-1035) with an optional URL path suffix (as specified by RFC-1738)
	ComponentName     ComponentName     `json:"componentName,omitempty"`
	IdentityAttribute IdentityAttribute `json:"identityAttribute,omitempty"`
	RelaxedSemver     RelaxedSemver     `json:"relaxedSemver,omitempty"`
	Component         Component         `json:"component,omitempty"`

	RepositoryContext  RepositoryContext  `json:"repositoryContext,omitempty"`
	Access             Access             `json:"access"`
	DigestSpec         DigestSpec         `json:"digestSpec,omitempty"`
	SignatureSpec      SignatureSpec      `json:"signatureSpec,omitempty"`
	Signature          Signature          `json:"signature,omitempty"`
	Source             Source             `json:"source,omitempty"`
	ComponentReference ComponentReference `json:"componentReference,omitempty"`
	Resource           Resource           `json:"resource,omitempty"`
}

type ComponentDescriptorStatus struct{}

// +kubebuilder:validation:MinLength=2
// +kubebuilder:validation:Pattern:=^[a-z0-9]([-_+a-z0-9]*[a-z0-9])?$
type IdentityAttributeKey string

// +kubebuilder:object:root=true
// +kubebuilder:storageversion

type ComponentDescriptor struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ComponentDescriptorSpec   `json:"spec,omitempty"`
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
