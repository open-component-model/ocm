package plugin

import (
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/descriptor"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/internal"
)

const (
	KIND_PLUGIN       = descriptor.KIND_PLUGIN
	KIND_UPLOADER     = descriptor.KIND_UPLOADER
	KIND_ACCESSMETHOD = descriptor.KIND_ACCESSMETHOD
	KIND_ACTION       = descriptor.KIND_ACTION
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
)
