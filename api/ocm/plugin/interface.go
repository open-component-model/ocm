package plugin

import (
	"ocm.software/ocm/api/ocm/plugin/descriptor"
	"ocm.software/ocm/api/ocm/plugin/internal"
	"ocm.software/ocm/api/tech/signing"
)

const (
	KIND_PLUGIN          = descriptor.KIND_PLUGIN
	KIND_UPLOADER        = descriptor.KIND_UPLOADER
	KIND_ACCESSMETHOD    = descriptor.KIND_ACCESSMETHOD
	KIND_ACTION          = descriptor.KIND_ACTION
	KIND_TRANSFERHANDLER = descriptor.KIND_TRANSFERHANDLER
)

var TAG = descriptor.REALM

type (
	Descriptor                  = descriptor.Descriptor
	ActionDescriptor            = descriptor.ActionDescriptor
	ValueMergeHandlerDescriptor = descriptor.ValueMergeHandlerDescriptor
	AccessMethodDescriptor      = descriptor.AccessMethodDescriptor
	DownloaderDescriptor        = descriptor.DownloaderDescriptor
	DownloaderKey               = descriptor.DownloaderKey
	UploaderDescriptor          = descriptor.UploaderDescriptor
	UploaderKey                 = descriptor.UploaderKey
	UploaderKeySet              = descriptor.UploaderKeySet
	ValueSetDefinition          = descriptor.ValueSetDefinition
	ValueSetDescriptor          = descriptor.ValueSetDescriptor
	CommandDescriptor           = descriptor.CommandDescriptor

	AccessSpecInfo       = internal.AccessSpecInfo
	UploadTargetSpecInfo = internal.UploadTargetSpecInfo

	SignatureSpec = internal.SignatureSpec
)

func SignatureSpecFor(sig *signing.Signature) *SignatureSpec {
	return internal.SignatureSpecFor(sig)
}

//
// Transfer handler types and constants
//

const (
	Q_UPDATE_VERSION    = internal.Q_UPDATE_VERSION
	Q_OVERWRITE_VERSION = internal.Q_OVERWRITE_VERSION
	Q_ENFORCE_TRANSPORT = internal.Q_ENFORCE_TRANSPORT
	Q_TRANSFER_VERSION  = internal.Q_TRANSFER_VERSION
	Q_TRANSFER_RESOURCE = internal.Q_TRANSFER_RESOURCE
	Q_TRANSFER_SOURCE   = internal.Q_TRANSFER_SOURCE
)

type (
	SourceComponentVersion = internal.SourceComponentVersion
	TargetRepositorySpec   = internal.TargetRepositorySpec
	TransferOptions        = internal.TransferOptions

	Artifact                   = internal.Artifact
	AccessInfo                 = internal.UniformAccessSpecInfo
	Question                   = internal.Question
	ComponentVersionQuestion   = internal.ComponentVersionQuestion
	ComponentReferenceQuestion = internal.ComponentReferenceQuestion
	ArtifactQuestion           = internal.ArtifactQuestion
	Resolution                 = internal.Resolution
	DecisionRequestResult      = internal.DecisionRequestResult
)
