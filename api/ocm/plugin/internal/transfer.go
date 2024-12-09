package internal

import (
	"encoding/json"

	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	v2 "ocm.software/ocm/api/ocm/compdesc/versions/v2"
	ocm "ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/tech"
)

const (
	Q_UPDATE_VERSION    = "updateversion"
	Q_OVERWRITE_VERSION = "overwriteversion"
	Q_ENFORCE_TRANSPORT = "enforcetransport"
	Q_TRANSFER_VERSION  = "transferversion"
	Q_TRANSFER_RESOURCE = "transferresource"
	Q_TRANSFER_SOURCE   = "transfersource"
)

type UniformAccessSpecInfo = tech.UniformAccessSpecInfo

type SourceComponentVersion struct {
	Name       string                    `json:"component"`
	Version    string                    `json:"version"`
	Provider   metav1.Provider           `json:"provider,omitempty"`
	Repository ocm.GenericRepositorySpec `json:"repository"`
	Labels     metav1.Labels             `json:"labels,omitempty"`
}

type TargetRepositorySpec = ocm.GenericRepositorySpec

type TransferOptions struct {
	Recursive         *bool            `json:"recursive,omitempty"`
	ResourcesByValue  *bool            `json:"resourcesByValue,omitempty"`
	LoalByValue       *bool            `json:"localByValue,omitempty"`
	SourcesByValue    *bool            `json:"sourcesByValue,omitempty"`
	KeepGlobalAccess  *bool            `json:"keepGlobalAccess,omitempty"`
	StopOnExisting    *bool            `json:"stopOnExisting,omitempty"`
	EnforceTransport  *bool            `json:"enforceTransport,omitempty"`
	Overwrite         *bool            `json:"overwrite,omitempty"`
	SkipUpdate        *bool            `json:"skipUpdate,omitempty"`
	OmitAccessTypes   []string         `json:"omitAccessTypes,omitempty"`
	OmitArtifactTypes []string         `json:"omitArtifactTypes,omitempty"`
	Special           *json.RawMessage `json:"special,omitempty"`
}

// Resolution describes the transport context for a component
// version, including the new handler specification and
// the source repository to use to look up the reference.
type Resolution struct {
	RepositorySpec *ocm.GenericRepositorySpec `json:"repository,omitempty"`
	// TransferHandler is the handler identity according to the transfer handler
	// name scheme.
	TransferHandler *string          `json:"transferHandler,omitempty"`
	HandlerOptions  *json.RawMessage `json:"handlerOptions,omitempty"`
}

type DecisionRequestResult struct {
	Error      string      `json:"error,omitempty"`
	Decision   bool        `json:"decision"`
	Resolution *Resolution `json:"resolution,omitempty"`
}

type ComponentVersionQuestion struct {
	Source  SourceComponentVersion `json:"source"`
	Target  TargetRepositorySpec   `json:"target"`
	Options TransferOptions        `json:"options"`
}

type ComponentReferenceQuestion struct {
	Source SourceComponentVersion `json:"source"`
	Target TargetRepositorySpec   `json:"target"`

	v2.ElementMeta `json:",inline"`
	ComponentName  string `json:"componentName"`

	Options TransferOptions `json:"options"`
}

type Artifact struct {
	Meta       v2.ElementMeta        `json:"metadata"`
	Access     ocm.GenericAccessSpec `json:"access"`
	AccessInfo UniformAccessSpecInfo `json:"accessInfo"`
}

type ArtifactQuestion struct {
	Source   SourceComponentVersion `json:"source"`
	Artifact Artifact               `json:"artifact"`
	Options  TransferOptions        `json:"options"`
}
