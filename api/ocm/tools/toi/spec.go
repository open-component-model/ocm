package toi

import (
	"encoding/json"
	"fmt"

	"github.com/mandelsoft/goutils/errors"
	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/credentials/cpi"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	resourcetypes "ocm.software/ocm/api/ocm/extensions/artifacttypes"
	common "ocm.software/ocm/api/utils/misc"
)

const (
	TypeTOIPackage               = "toiPackage"
	PackageSpecificationMimeType = "application/vnd.toi.ocm.software.package.v1+yaml"

	TypeYAML = resourcetypes.OCM_YAML

	AdditionalResourceConfigFile      = "configFile"
	AdditionalResourceCredentialsFile = "credentialsFile"
)

const (
	TypeTOIExecutor               = "toiExecutor"
	ExecutorSpecificationMimeType = "application/vnd.toi.ocm.software.executor.v1+yaml"
)

type PackageSpecification struct {
	CredentialsRequest  `json:",inline"`
	Template            json.RawMessage                `json:"configTemplate,omitempty"`
	Libraries           []metav1.ResourceReference     `json:"templateLibraries,omitempty"`
	Scheme              json.RawMessage                `json:"configScheme,omitempty"`
	Executors           []Executor                     `json:"executors"`
	Description         string                         `json:"description"`
	AdditionalResources map[string]*AdditionalResource `json:"additionalResources,omitempty"`
}

type AdditionalResource struct {
	*metav1.ResourceReference `json:",inline"`
	Content                   json.RawMessage `json:"content,omitempty"`
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

func (e *Executor) Name() string {
	if e.ResourceRef != nil {
		return e.ResourceRef.String()
	}
	if e.Image != nil {
		return e.Image.String()
	}
	return "unspecified executor"
}

type Image struct {
	Ref    string `json:"ref"`
	Digest string `json:"digest"`
}

func (i *Image) String() string {
	r := "<noref>"
	if i.Ref != "" {
		r = i.Ref
	}
	if i.Digest != "" {
		r += "@" + i.Digest
	}
	return r
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

////////////////////////////////////////////////////////////////////////////////

type CredentialsRequest struct {
	Credentials map[string]CredentialsRequestSpec `json:"credentials,omitempty"`
}

type CredentialsRequestSpec struct {
	// ConsumerId specified to consumer id the credentials are used for
	ConsumerId credentials.ConsumerIdentity `json:"consumerId,omitempty"`
	// Description described the usecase the credentials will be used for
	Description string `json:"description"`
	// Properties describes the meaning of the used properties for this
	// credential set.
	Properties common.Properties `json:"properties"`
	// Optional set to true make the request optional
	Optional bool `json:"optional,omitempty"`
}

var ErrUndefined error = errors.New("nil reference")

func (s *CredentialsRequestSpec) Match(o *CredentialsRequestSpec) error {
	if o == nil {
		return ErrUndefined
	}
	if !s.ConsumerId.Equals(o.ConsumerId) {
		return fmt.Errorf("consumer id mismatch")
	}
	for k := range o.Properties {
		if _, ok := s.Properties[k]; !ok {
			return fmt.Errorf("property %q not declared", k)
		}
	}
	if s.Optional && !o.Optional {
		return fmt.Errorf("cannot be optional")
	}
	return nil
}

type Credentials struct {
	Credentials map[string]CredentialSpec `json:"credentials,omitempty"`

	// Forwarded may define a list of consumer ids, which should be taken from the
	// local configuration and forwarded to the TOI executor in addition to the
	// credentials explicitly requested by the installation package.
	Forwarded []ForwardSpec `json:"forwardedConsumers,omitempty"`
}

type CredentialSpec struct {
	// ConsumerId specifies the consumer id to look for the credentials
	ConsumerId credentials.ConsumerIdentity `json:"consumerId,omitempty"`
	// ConsumerType is the optional type used for matching the credentials
	ConsumerType string `json:"consumerType,omitempty"`
	// Reference refers to credentials store in some other repo
	Reference *cpi.GenericCredentialsSpec `json:"reference,omitempty"`
	// Credentials are direct credentials (one of Reference or Credentials must be set)
	Credentials common.Properties `json:"credentials,omitempty"`

	// TargetConsumerId specifies the consumer id to feed with these credentials
	TargetConsumerId credentials.ConsumerIdentity `json:"targetConsumerId,omitempty"`
}

type ForwardSpec struct {
	// ConsumerId specifies the consumer id to look for the credentials
	ConsumerId credentials.ConsumerIdentity `json:"consumerId"`
	// ConsumerType is the optional type used for matching the credentials
	ConsumerType string `json:"consumerType,omitempty"`
}
