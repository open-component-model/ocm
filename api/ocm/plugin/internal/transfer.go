package internal

import (
	"encoding/json"
	"reflect"

	"github.com/mandelsoft/goutils/generics"

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

var TransferHandlerQuestions = map[string]reflect.Type{
	Q_UPDATE_VERSION:    generics.TypeOf[ComponentVersionQuestionArguments](),
	Q_ENFORCE_TRANSPORT: generics.TypeOf[ComponentVersionQuestionArguments](),
	Q_OVERWRITE_VERSION: generics.TypeOf[ComponentVersionQuestionArguments](),
	Q_TRANSFER_VERSION:  generics.TypeOf[ComponentReferenceQuestionArguments](),
	Q_TRANSFER_RESOURCE: generics.TypeOf[ArtifactQuestionArguments](),
	Q_TRANSFER_SOURCE:   generics.TypeOf[ArtifactQuestionArguments](),
}

type UniformAccessSpecInfo = tech.UniformAccessSpecInfo

type SourceComponentVersion struct {
	Name       string                    `json:"component"`
	Version    string                    `json:"version"`
	Provider   metav1.Provider           `json:"provider,omitempty"`
	Repository ocm.GenericRepositorySpec `json:"repository"`
	Labels     metav1.Labels             `json:"labels,omitempty"`
}

type TargetRepositorySpec = ocm.GenericRepositorySpec

// TransferOptions are the standard transfer options from
// the standard transfer handler.
// Like other transfer handler types it is possible to define
// more options. Those non-standard options are passed
// via a json.RawMessage in the field special.
type TransferOptions struct {
	Recursive         *bool            `json:"recursive,omitempty"`
	ResourcesByValue  *bool            `json:"resourcesByValue,omitempty"`
	LocalByValue      *bool            `json:"localByValue,omitempty"`
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
	TransferHandler string `json:"transferHandler,omitempty"`
	// TransferOptions may describe modified options used for sub-sequent
	// transfers.
	TransferOptions *TransferOptions `json:"transferOptions,omitempty"`
}

// DecisionRequestResult is the structure of the answer
// the plugin has to return for a question.
type DecisionRequestResult struct {
	Error      string      `json:"error,omitempty"`
	Decision   bool        `json:"decision"`
	Resolution *Resolution `json:"resolution,omitempty"`
}

// QuestionArguments is the interface for the question attributes
// differing for the various questions types.
// There are three basic attribute sets:
//   - ComponentVersionQuestionArguments
//   - ComponentReferenceQuestionArguments
//   - ArtifactQuestionArguments
//
// For type assignments see TransferHandlerQuestions.
type QuestionArguments interface {
	QuestionArgumentsType() string
}

// ComponentVersionQuestionArguments describes the question arguments
// given for a component version related question.
type ComponentVersionQuestionArguments struct {
	Source  SourceComponentVersion `json:"source"`
	Target  TargetRepositorySpec   `json:"target"`
	Options TransferOptions        `json:"options"`
}

var _ QuestionArguments = (*ComponentVersionQuestionArguments)(nil)

func (a *ComponentVersionQuestionArguments) QuestionArgumentsType() string {
	return "ComponentVersionQuestionArguments"
}

// ComponentReferenceQuestionArguments  describes the question arguments
// given for a component version reference related question.
type ComponentReferenceQuestionArguments struct {
	Source SourceComponentVersion `json:"source"`
	Target TargetRepositorySpec   `json:"target"`

	v2.ElementMeta `json:",inline"`
	ComponentName  string `json:"componentName"`

	Options TransferOptions `json:"options"`
}

var _ QuestionArguments = (*ComponentReferenceQuestionArguments)(nil)

func (a *ComponentReferenceQuestionArguments) QuestionArgumentsType() string {
	return "ComponentReferenceQuestionArguments"
}

type Artifact struct {
	Meta       v2.ElementMeta        `json:"metadata"`
	Access     ocm.GenericAccessSpec `json:"access"`
	AccessInfo UniformAccessSpecInfo `json:"accessInfo"`
}

// ArtifactQuestionArguments  describes the question arguments
// given for an artifact related question.
type ArtifactQuestionArguments struct {
	Source   SourceComponentVersion `json:"source"`
	Artifact Artifact               `json:"artifact"`
	Options  TransferOptions        `json:"options"`
}

var _ QuestionArguments = (*ArtifactQuestionArguments)(nil)

func (a *ArtifactQuestionArguments) QuestionArgumentsType() string {
	return "ArtifactQuestionArguments"
}
