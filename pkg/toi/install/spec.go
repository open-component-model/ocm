// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package install

import (
	"encoding/json"

	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
)

const (
	TypeTOIPackage               = "toiPackage"
	PackageSpecificationMimeType = "application/vnd.toi.ocm.software.package.v1+yaml"
)

const (
	TypeTOIExecutor               = "toiExecutor"
	ExecutorSpecificationMimeType = "application/vnd.toi.ocm.software.executor.v1+yaml"
)

type PackageSpecification struct {
	CredentialsRequest `json:",inline"`
	Template           json.RawMessage            `json:"configTemplate,omitempty"`
	Libraries          []metav1.ResourceReference `json:"templateLibraries,omitempty"`
	Scheme             json.RawMessage            `json:"configScheme,omitempty"`
	Executors          []Executor                 `json:"executors"`
}

type Executor struct {
	Actions           []string                  `json:"actions,omitempty"`
	ResourceRef       *metav1.ResourceReference `json:"resourceRef,omitempty"`
	Image             *Image                    `json:"image,omitempty"`
	CredentialMapping map[string]string         `json:"credentialMapping,omitempty"`
	ParameterMapping  json.RawMessage           `json:"parameterMapping,omitempty"`
	Config            json.RawMessage           `json:"config,omitempty"`
	Outputs           map[string]string         `json:"outputs,omitempty"`
}

type Image struct {
	Ref    string `json:"ref"`
	Digest string `json:"digest"`
}

////////////////////////////////////////////////////////////////////////////////

type ExecutorSpecification struct {
	CredentialsRequest `json:",inline"`
	Actions            []string                   `json:"actions,omitempty"`
	Image              *Image                     `json:"image,omitempty"`
	ImageRef           *metav1.ResourceReference  `json:"imageRef,omitempty"`
	Template           json.RawMessage            `json:"configTemplate,omitempty"`
	Libraries          []metav1.ResourceReference `json:"templateLibraries,omitempty"`
	Scheme             json.RawMessage            `json:"configScheme,omitempty"`
	Outputs            map[string]OutputSpec      `json:"outputs,omitempty"`
}

type OutputSpec struct {
	Description string `json:"description,omitempty"`
}
