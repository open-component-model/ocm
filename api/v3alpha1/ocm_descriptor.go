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

package v3alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TODO: Think about how to deal with this object definition in the JSON schema.
type Object struct {
}

type Provider struct {
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	Labels []Label `json:"labels"`
}

// Meta is component version metadata
type Meta struct {
	// +kubebuilder:validation:Required
	Name ComponentName `json:"name"`
	// +kubebuilder:validation:Required
	Version RelaxedSemver `json:"version"`

	Labels   []Label  `json:"labels"`
	Provider Provider `json:"provider"`
}

type Label struct {
	// +kubebuilder:validation:Required
	Name string `json:"name,omitempty"`
	// +kubebuilder:validation:Required
	Values string `json:"values,omitempty"`
	// +kubebuilder:validation:Pattern:=`^v[0-9]+$`
	Version   string `json:"version"`
	Signature bool   `json:"signature"`
}

type IdentityAttribute struct {
	IdentityAttributeKeys []IdentityAttributeKey `json:"identityAttributeKeys,omitempty"`
}

type RepositoryContext struct {
	// +kubebuilder:validation:Required
	Type string `json:"type"`
}

type OCIRepositoryContext struct {
	// +kubebuilder:validation:Required
	BaseURL           string            `json:"baseURL,omitempty"`
	RepositoryContext RepositoryContext `json:"repositoryContext"`

	// +kubebuilder:validation:Enum:=urlPath;sha256-digest
	ComponentNameMapping string `json:"componentNameMapping,omitempty"`
	// +kubebuilder:validation:Enum:=ociRegistry;OCIRegistry
	Type string `json:"type,omitempty"`
}

type Access struct {
	// Type is the base type for the access of a source or resource
	// +kubebuilder:validation:Required
	Type string `json:"type"`
}

type GithubAccess struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum:=github
	Type    string `json:"type"`
	RepoUrl string `json:"repoUrl"`
	Ref     string `json:"ref"`
	Commit  string `json:"commit"`
}

type NoneAccess struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum:=None
	Type string `json:"type"`
}

type HTTPAccess struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum:=http
	Type string `json:"type"`
	URL  string `json:"url"`
}

type GenericAccess struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum:=generic
	Type string `json:"type"`
}

type OCIImageAccess struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum:=ociRegistry
	Type           string `json:"type"`
	ImageReference string `json:"imageReference"`
}

type OCIBlobAccess struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum:=ociBlob
	Type string `json:"type"`
	// Ref is an oci reference to the manifest.
	Ref string `json:"ref"`
	// MediaType The media type of the object this access refers to.
	MediaType string `json:"mediaType"`
	// Digest The digest of the targeted content.
	Digest string `json:"digest"`
	// Size The size in bytes of the blob.
	Size int `json:"size"`
}

type LocalFilesystemBlobAccess struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum:=localFilesystemBlob
	Type string `json:"type"`
	// Filename is the filename of the blob that is located in the "blobs" directory.
	Filename string `json:"filename"`
}

type LocalOciBlobAccess struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum:=localOciBlob
	Type string `json:"type"`
	// Digest of the layer within the current component descriptor
	Digest string `json:"digest"`
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

type OCIImageResource struct {
	Name    IdentityAttributeKey `json:"name"`
	Version RelaxedSemver        `json:"version"`
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum:=ociImage
	Type   string         `json:"type"`
	Labels []Label        `json:"labels,omitempty"`
	Access OCIImageAccess `json:"access"`

	ExtraIdentity IdentityAttribute `json:"extraIdentity,omitempty"`
	Digest        DigestSpec        `json:"digest"`
}

type SourceDefinition struct {
	// +kubebuilder:validation:Required
	Name IdentityAttributeKey `json:"name"`
	// +kubebuilder:validation:Required
	Version RelaxedSemver `json:"version"`
	// +kubebuilder:validation:Required
	Type string `json:"type"`
	// TODO: Access can be either of type Access or githubAccess or httpAccess. We can do this by making this
	// field use *apiextensions.JSON which validates any valid JSON. Then, we can unmarshall it and
	// use the values and not loose type safety. Or, have an explicit field called GithubAcess and HTTPAccess
	// which can either be provided or not.
	// +kubebuilder:validation:Required
	Access Access `json:"access,omitempty"`

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
	IdentitySelector IdentityAttribute `json:"identitySelector,omitempty"`
	Labels           []Label           `json:"labels,omitempty"`
}

// ResourceType is the base type for resources.
type ResourceType struct {
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
	Access Access `json:"access,omitempty"`

	SrcRefs       []SourceReferences `json:"srcRefs,omitempty"`
	ExtraIdentity IdentityAttribute  `json:"extraIdentity,omitempty"`
	Labels        []Label            `json:"labels,omitempty"`
	Digest        DigestSpec         `json:"digest,omitempty"`
}

type GenericResource struct {
	Name    IdentityAttributeKey `json:"name"`
	Version RelaxedSemver        `json:"version"`
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum:=generic
	Type   string        `json:"type"`
	Access GenericAccess `json:"access"`
	Digest DigestSpec    `json:"digest,omitempty"`

	Labels []Label `json:"labels,omitempty"`
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
	RepositoryContexts []OCIRepositoryContext `json:"repositoryContexts,omitempty"`
	// +kubebuilder:validation:Required
	Provider string `json:"provider,omitempty"`
	// +kubebuilder:validation:Required
	Sources []SourceDefinition `json:"sources,omitempty"`
	// +kubebuilder:validation:Required
	ComponentReferences []ComponentReference `json:"componentReferences,omitempty"`
	// +kubebuilder:validation:Required
	// TODO: This can be multiple types. So we need a JSON schema here too. Unless we define all of the accesses
	// as a separate field. But I believe that defeats the purpose of the spec.
	Resources []ResourceType `json:"resources,omitempty"`

	Labels []Label `json:"labels,omitempty"`
}

type ComponentDescriptorSpec struct {
	Meta  Meta    `json:"meta,omitempty"`
	Label []Label `json:"label,omitempty"`
	// ComponentName MUST start with a valid domain name (as specified by RFC-1034, RFC-1035) with an optional URL path suffix (as specified by RFC-1738)
	// +kubebuilder:validation:Pattern:=`^[a-z][-a-z0-9]*([.][a-z][-a-z0-9]*)*[.][a-z]{2,}(/[a-z][-a-z0-9_]*([.][a-z][-a-z0-9_]*)*)+$`
	ComponentName     string            `json:"componentName,omitempty"`
	IdentityAttribute IdentityAttribute `json:"identityAttribute,omitempty"`
	RelaxedSemver     RelaxedSemver     `json:"relaxedSemver,omitempty"`
	Component         Component         `json:"component,omitempty"`

	RepositoryContext  RepositoryContext  `json:"repositoryContext,omitempty"`
	Access             Access             `json:"access"`
	DigestSpec         DigestSpec         `json:"digestSpec,omitempty"`
	SignatureSpec      SignatureSpec      `json:"signatureSpec,omitempty"`
	Signature          Signature          `json:"signature,omitempty"`
	Source             SourceDefinition   `json:"source,omitempty"`
	ComponentReference ComponentReference `json:"componentReference,omitempty"`
	Resource           ResourceType       `json:"resource,omitempty"`
}

type ComponentDescriptorStatus struct{}

// +kubebuilder:validation:MinLength=2
// +kubebuilder:validation:Pattern:=`^[a-z0-9]([-_+a-z0-9]*[a-z0-9])?$`
type IdentityAttributeKey string

// +kubebuilder:object:root=true

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
