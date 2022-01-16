// Copyright 2020 Copyright (c) 2020 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package compdesc

import (
	"encoding/json"
	"fmt"

	metav1 "github.com/gardener/ocm/pkg/ocm/compdesc/meta/v1"
)

const DefaultSchemeVersion = "v2"

type ComponentDescriptorVersion interface {
	GetVersion() string
	Decode(data []byte, opts *DecodeOptions) (interface{}, error)
	ConvertFrom(desc *ComponentDescriptor) (interface{}, error)
	ConvertTo(interface{}) (*ComponentDescriptor, error)
}

type ComponentDescriptorVersions map[string]ComponentDescriptorVersion

func (v ComponentDescriptorVersions) Register(scheme ComponentDescriptorVersion) {
	v[scheme.GetVersion()] = scheme
}

var DefaultSchemes = ComponentDescriptorVersions{}

func RegisterScheme(scheme ComponentDescriptorVersion) {
	DefaultSchemes.Register(scheme)
}

////////////////////////////////////////////////////////////////////////////////

// Decode decodes a component into the given object.
func Decode(data []byte, opts ...DecodeOption) (*ComponentDescriptor, error) {
	o := &DecodeOptions{Codec: DefaultYAMLCodec}
	o.ApplyOptions(opts)

	raw := make(map[string]json.RawMessage)
	if err := o.Codec.Decode(data, &raw); err != nil {
		return nil, err
	}

	var metadata metav1.Metadata
	if err := o.Codec.Decode(raw["meta"], &metadata); err != nil {
		return nil, err
	}

	version := DefaultSchemes[metadata.Version]
	if version == nil {
		return nil, fmt.Errorf("unsupported schema version %q", metadata.Version)
	}

	versioned, err := version.Decode(data, o)
	if err != nil {
		return nil, err
	}
	return version.ConvertTo(versioned)
}

// DecodeOptions defines decode options for the codec.
type DecodeOptions struct {
	Codec             Codec
	DisableValidation bool
	StrictMode        bool
}

var _ DecodeOption = &DecodeOptions{}

// ApplyDecodeOption applies the actual options.
func (o *DecodeOptions) ApplyDecodeOption(options *DecodeOptions) {
	if o == nil {
		return
	}
	if o.Codec != nil {
		options.Codec = o.Codec
	}
	options.DisableValidation = o.DisableValidation
	options.StrictMode = o.StrictMode
}

// ApplyOptions applies the given list options on these options,
// and then returns itself (for convenient chaining).
func (o *DecodeOptions) ApplyOptions(opts []DecodeOption) *DecodeOptions {
	for _, opt := range opts {
		if opt != nil {
			opt.ApplyDecodeOption(o)
		}
	}
	return o
}

// DecodeOption is the interface to specify different cache options
type DecodeOption interface {
	ApplyDecodeOption(options *DecodeOptions)
}

// StrictMode enables or disables strict mode parsing.
type StrictMode bool

// ApplyDecodeOption applies the configured strict mode.
func (s StrictMode) ApplyDecodeOption(options *DecodeOptions) {
	options.StrictMode = bool(s)
}

// DisableValidation enables or disables validation of the component descriptor.
type DisableValidation bool

// ApplyDecodeOption applies the validation disable option.
func (v DisableValidation) ApplyDecodeOption(options *DecodeOptions) {
	options.DisableValidation = bool(v)
}

////////////////////////////////////////////////////////////////////////////////

// Encode encodes a component into the given object.
// The obj is expected to be of type v2.ComponentDescriptor or v2.ComponentDescriptorList.
// If the serialization version is left blank, the schema version configured in the
// component descriptor will be used.
func Encode(obj *ComponentDescriptor, opts ...EncodeOption) ([]byte, error) {
	o := (&EncodeOptions{}).ApplyOptions(opts).DefaultFor(obj)
	cv := DefaultSchemes[o.SchemaVersion]
	if cv == nil {
		if cv == nil {
			return nil, fmt.Errorf("unsupported schema version %q", o.SchemaVersion)
		}
	}
	v, err := cv.ConvertFrom(obj)
	if err != nil {
		return nil, err
	}
	return o.Codec.Encode(v)
}

////////////////////////////////////////////////////////////////////////////////

type EncodeOptions struct {
	Codec         Codec
	SchemaVersion string
}

var _ EncodeOption = &EncodeOptions{}

// ApplyDecodeOption applies the actual options.
func (o *EncodeOptions) ApplyEncodeOption(options *EncodeOptions) {
	if o == nil {
		return
	}
	if o.Codec != nil {
		options.Codec = o.Codec
	}
	if o.SchemaVersion != "" {
		options.SchemaVersion = o.SchemaVersion
	}
}

func (o *EncodeOptions) DefaultFor(cd *ComponentDescriptor) *EncodeOptions {
	if o.Codec == nil {
		o.Codec = DefaultYAMLCodec
	}
	if o.SchemaVersion == "" {
		o.SchemaVersion = cd.Metadata.ConfiguredVersion
	}
	if o.SchemaVersion == "" {
		o.SchemaVersion = DefaultSchemeVersion
	}
	return o
}

// ApplyOptions applies the given list options on these options,
// and then returns itself (for convenient chaining).
func (o *EncodeOptions) ApplyOptions(opts []EncodeOption) *EncodeOptions {
	for _, opt := range opts {
		if opt != nil {
			opt.ApplyEncodeOption(o)
		}
	}
	return o
}

// EncodeOption is the interface to specify different encode options
type EncodeOption interface {
	ApplyEncodeOption(options *EncodeOptions)
}

// SchemaVersion enforces a dedicated schema version .
type SchemaVersion string

// ApplyEncodeOption applies the configured schema version.
func (o SchemaVersion) ApplyEncodeOption(options *EncodeOptions) {
	options.SchemaVersion = string(o)
}

// CodecWrappers can be used as EncodeOption, also
