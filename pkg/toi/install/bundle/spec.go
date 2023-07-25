// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package bundle

import (
	"encoding/json"

	metav1 "github.com/open-component-model/ocm/v2/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/v2/pkg/toi"
	"github.com/open-component-model/ocm/v2/pkg/toi/install"
)

type BundleSpecification struct {
	install.CredentialsRequest `json:",inline"`
	Template                   json.RawMessage            `json:"configTemplate,omitempty"`
	Libraries                  []metav1.ResourceReference `json:"templateLibraries,omitempty"`
	Scheme                     json.RawMessage            `json:"configScheme,omitempty"`

	Actions []string          `json:"actions,omitempty"`
	Outputs map[string]string `json:"outputs,omitempty"`
}

type InstallationSpecification struct {
	ResourceRef *metav1.ResourceReference `json:"resourceRef,omitempty"`
	Image       *toi.Image                `json:"image,omitempty"`

	Actions  map[string]string `json:"actions,omitempty"`
	Required map[string]string `json:"required,omitempty"`
}

type InstallationValues struct {
	install.Credentials `json:",inline"`
	Settings            json.RawMessage `json:"values,omitempty"`
}
