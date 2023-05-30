// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package ppi

import (
	"encoding/json"
	"io"

	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext/action"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/options"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/descriptor"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/internal"
	"github.com/open-component-model/ocm/pkg/runtime"
)

type (
	Descriptor             = descriptor.Descriptor
	UploaderKey            = descriptor.UploaderKey
	UploaderDescriptor     = descriptor.UploaderDescriptor
	DownloaderKey          = descriptor.DownloaderKey
	DownloaderDescriptor   = descriptor.DownloaderDescriptor
	AccessMethodDescriptor = descriptor.AccessMethodDescriptor
	CLIOption              = descriptor.CLIOption

	ActionSpecInfo       = internal.ActionSpecInfo
	AccessSpecInfo       = internal.AccessSpecInfo
	UploadTargetSpecInfo = internal.UploadTargetSpecInfo
)

var REALM = descriptor.REALM

type Plugin interface {
	Name() string
	Version() string
	Descriptor() descriptor.Descriptor

	SetDescriptorTweaker(func(descriptor descriptor.Descriptor) descriptor.Descriptor)

	SetShort(s string)
	SetLong(s string)
	SetConfigParser(config func(raw json.RawMessage) (interface{}, error))

	RegisterDownloader(arttype, mediatype string, u Downloader) error
	GetDownloader(name string) Downloader
	GetDownloaderFor(arttype, mediatype string) Downloader

	RegisterUploader(arttype, mediatype string, u Uploader) error
	GetUploader(name string) Uploader
	GetUploaderFor(arttype, mediatype string) Uploader
	DecodeUploadTargetSpecification(data []byte) (UploadTargetSpec, error)

	RegisterAccessMethod(m AccessMethod) error
	DecodeAccessSpecification(data []byte) (AccessSpec, error)
	GetAccessMethod(name string, version string) AccessMethod

	RegisterAction(a Action) error
	DecodeAction(data []byte) (ActionSpec, error)
	GetAction(name string) Action

	GetOptions() *Options
	GetConfig() (interface{}, error)
}

type AccessMethod interface {
	runtime.TypedObjectDecoder[AccessSpec]

	Name() string
	Version() string

	// Options provides the list of CLI options supported to compose the access
	// specification.
	Options() []options.OptionType

	// Description provides a general description for the access mehod kind.
	Description() string
	// Format describes the attributes of the dedicated version.
	Format() string

	ValidateSpecification(p Plugin, spec AccessSpec) (info *AccessSpecInfo, err error)
	Reader(p Plugin, spec AccessSpec, creds credentials.Credentials) (io.ReadCloser, error)
	ComposeAccessSpecification(p Plugin, opts Config, config Config) error
}

type AccessSpec = runtime.TypedObject

type AccessSpecProvider func() AccessSpec

type UploadFormats runtime.KnownTypes[runtime.TypedObject, runtime.TypedObjectDecoder[runtime.TypedObject]]

type Uploader interface {
	Decoders() UploadFormats

	Name() string
	Description() string

	ValidateSpecification(p Plugin, spec UploadTargetSpec) (info *UploadTargetSpecInfo, err error)
	Writer(p Plugin, arttype, mediatype string, hint string, spec UploadTargetSpec, creds credentials.Credentials) (io.WriteCloser, AccessSpecProvider, error)
}

type UploadTargetSpec = runtime.TypedObject

type DownloadResultProvider func() (string, error)

type Downloader interface {
	Name() string
	Description() string
	ConfigSchema() []byte

	Writer(p Plugin, arttype, mediatype string, filepath string, config []byte) (io.WriteCloser, DownloadResultProvider, error)
}

type ActionSpec = action.ActionSpec

type ActionResult = action.ActionResult

type Action interface {
	Name() string
	Description() string
	DefaultSelectors() []string
	ConsumerType() string

	Execute(p Plugin, spec ActionSpec, creds credentials.DirectCredentials) (result ActionResult, err error)
}
