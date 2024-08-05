package bundle

import (
	"encoding/json"

	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/tools/toi"
	"ocm.software/ocm/api/ocm/tools/toi/install"
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
