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

// ApplyOptions applies the given list options on these options,
// and then returns itself (for convenient chaining).
func (o *DecodeOptions) ApplyOptions(opts []DecodeOption) *DecodeOptions {
	for _, opt := range opts {
		if opt != nil {
			opt.ApplyOption(o)
		}
	}
	return o
}

// DecodeOption is the interface to specify different cache options
type DecodeOption interface {
	ApplyOption(options *DecodeOptions)
}

// StrictMode enables or disables strict mode parsing.
type StrictMode bool

// ApplyOption applies the configured strict mode.
func (s StrictMode) ApplyOption(options *DecodeOptions) {
	options.StrictMode = bool(s)
}

// DisableValidation enables or disables validation of the component descriptor.
type DisableValidation bool

// ApplyOption applies the validation disable option.
func (v DisableValidation) ApplyOption(options *DecodeOptions) {
	options.DisableValidation = bool(v)
}

////////////////////////////////////////////////////////////////////////////////

// Encode encodes a component into the given object.
// The obj is expected to be of type v2.ComponentDescriptor or v2.ComponentDescriptorList.
// If the serialization version is left blank, the schema version configured in the
// component descriptor will be used.
func Encode(obj *ComponentDescriptor, version string, codec Codec) ([]byte, error) {
	if version == "" {
		version = obj.Metadata.ConfiguredVersion
	}
	cv := DefaultSchemes[version]
	if cv == nil {
		if cv == nil {
			return nil, fmt.Errorf("unsupported schema version %q", version)
		}
	}
	v, err := cv.ConvertFrom(obj)
	if err != nil {
		return nil, err
	}
	return codec.Encode(v)
}
