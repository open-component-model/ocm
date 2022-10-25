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
	Descriptor     = internal.Descriptor
	AccessSpecInfo = internal.AccessSpecInfo
)

const KIND_PLUGIN = "plugin"

type Plugin interface {
	Name() string
	Version() string
	Descriptor() internal.Descriptor

	RegisterAccessMethod(m AccessMethod) error
	DecodeAccessSpecification(data []byte) (AccessSpec, error)
	GetAccessMethod(name string, version string) AccessMethod

	Options() *Options
}

type AccessMethod interface {
	runtime.TypedObjectDecoder

	Name() string
	Version() string

	ValidateSpecification(p Plugin, spec AccessSpec) (info *AccessSpecInfo, err error)
	Reader(p Plugin, spec AccessSpec, creds credentials.Credentials) (io.ReadCloser, error)
	Writer(p Plugin, mediatype string, creds credentials.Credentials) (io.WriteCloser, AccessSpecProvider, error)
}
