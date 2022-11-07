// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package ppi

import (
	"io"

	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/internal"
	"github.com/open-component-model/ocm/pkg/runtime"
)

type (
	Descriptor             = internal.Descriptor
	AccessSpecInfo         = internal.AccessSpecInfo
	UploadTargetSpecInfo   = internal.UploadTargetSpecInfo
	UploaderKey            = internal.UploaderKey
	UploaderDescriptor     = internal.UploaderDescriptor
	DownloaderKey          = internal.DownloaderKey
	DownloaderDescriptor   = internal.DownloaderDescriptor
	AccessMethodDescriptor = internal.AccessMethodDescriptor
)

const (
	KIND_PLUGIN       = internal.KIND_PLUGIN
	KIND_DOWNLOADER   = internal.KIND_DOWNLOADER
	KIND_UPLOADER     = internal.KIND_UPLOADER
	KIND_ACCESSMETHOD = internal.KIND_ACCESSMETHOD
)

var TAG = internal.TAG

type Plugin interface {
	Name() string
	Version() string
	Descriptor() internal.Descriptor

	SetShort(s string)
	SetLong(s string)

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

	Options() *Options
}

type AccessMethod interface {
	runtime.TypedObjectDecoder

	Name() string
	Version() string

	// Description provides a general description for the access mehod kind.
	Description() string
	// Format describes the attributes of the dedicated version.
	Format() string

	ValidateSpecification(p Plugin, spec AccessSpec) (info *AccessSpecInfo, err error)
	Reader(p Plugin, spec AccessSpec, creds credentials.Credentials) (io.ReadCloser, error)
}

type AccessSpec runtime.VersionedTypedObject

type AccessSpecProvider func() AccessSpec

type Uploader interface {
	Decoders() map[string]runtime.TypedObjectDecoder

	Name() string
	Description() string

	ValidateSpecification(p Plugin, spec UploadTargetSpec) (info *UploadTargetSpecInfo, err error)
	Writer(p Plugin, arttype, mediatype string, hint string, spec UploadTargetSpec, creds credentials.Credentials) (io.WriteCloser, AccessSpecProvider, error)
}

type UploadTargetSpec runtime.VersionedTypedObject

type DownloadResultProvider func() (string, error)

type Downloader interface {
	Name() string
	Description() string

	Writer(p Plugin, arttype, mediatype string, filepath string) (io.WriteCloser, DownloadResultProvider, error)
}
