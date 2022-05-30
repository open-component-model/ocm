// Copyright 2022 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package signing

import (
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/signing"
	"github.com/open-component-model/ocm/pkg/signing/handlers/rsa"
	"github.com/open-component-model/ocm/pkg/signing/hasher/sha256"
)

type Option interface {
	ApplySigningOption(o *Options)
}

////////////////////////////////////////////////////////////////////////////////

type recursive struct {
	flag bool
}

func Recursive(flag bool) Option {
	return &recursive{flag}
}

func (o *recursive) ApplySigningOption(opts *Options) {
	opts.Recursively = o.flag
}

////////////////////////////////////////////////////////////////////////////////

type update struct {
	flag bool
}

func Update(flag bool) Option {
	return &update{flag}
}

func (o *update) ApplySigningOption(opts *Options) {
	opts.Update = o.flag
}

////////////////////////////////////////////////////////////////////////////////

type signer struct {
	signer signing.Signer
}

func Signer(h signing.Signer) Option {
	return &signer{h}
}

func (o *signer) ApplySigningOption(opts *Options) {
	opts.Signer = o.signer
	if o.signer != nil {
		if v, ok := o.signer.(signing.Verifier); ok {
			opts.Verifier = v
		}
	}
}

////////////////////////////////////////////////////////////////////////////////

type verifier struct {
	verifier signing.Verifier
}

func Verifier(h signing.Verifier) Option {
	return &verifier{h}
}

func (o *verifier) ApplySigningOption(opts *Options) {
	opts.Verifier = o.verifier
	if o.verifier != nil {
		if s, ok := o.verifier.(signing.Signer); ok {
			opts.Signer = s
		}
	}
}

////////////////////////////////////////////////////////////////////////////////

type resolver struct {
	resolver ocm.ComponentVersionResolver
}

func Resolver(h ocm.ComponentVersionResolver) Option {
	return &resolver{h}
}

func (o *resolver) ApplySigningOption(opts *Options) {
	opts.Resolver = o.resolver
}

////////////////////////////////////////////////////////////////////////////////

type registry struct {
	registry signing.Registry
}

func Registry(h signing.Registry) Option {
	return &registry{h}
}

func (o *registry) ApplySigningOption(opts *Options) {
	opts.Registry = o.registry
}

////////////////////////////////////////////////////////////////////////////////

type Options struct {
	Update          bool
	Recursively     bool
	Signer          signing.Signer
	Verifier        signing.Verifier
	Hasher          signing.Hasher
	Registry        signing.Registry
	Resolver        ocm.ComponentVersionResolver
	SkipAccessTypes map[string]bool
	SignatureName   string
}

var _ Option = (*Options)(nil)

func (o *Options) ApplySigningOption(opts *Options) {
	if o.Signer != nil {
		opts.Signer = o.Signer
	}
	if o.Verifier != nil {
		opts.Verifier = o.Verifier
	}
	if o.Hasher != nil {
		opts.Hasher = o.Hasher
	}
	opts.Recursively = o.Recursively
	opts.Update = o.Update
}

func (o *Options) Default(registry signing.Registry) {
	if o.Registry == nil {
		o.Registry = registry
	} else {
		registry = o.Registry
	}
	if o.SkipAccessTypes == nil {
		o.SkipAccessTypes = map[string]bool{}
	}
	if o.Signer == nil {
		o.Signer = registry.GetSigner(rsa.Algorithm)
	}
	if o.Verifier == nil {
		if o.Signer != nil {
			if v, ok := o.Signer.(signing.Verifier); ok {
				o.Verifier = v
			}
		}
		if o.Verifier == nil {
			o.Verifier = registry.GetVerifier(rsa.Algorithm)
		}
	}
	if o.Hasher == nil {
		o.Hasher = registry.GetHasher(sha256.Algorithm)
	}
	if o.Registry == nil {
		o.Registry = signing.DefaultRegistry()
	}
}

func (o *Options) For(digest *metav1.DigestSpec) (*Options, error) {
	if digest == nil {
		return o, nil
	}
	opts := *o
	opts.Hasher = o.Registry.GetHasher(digest.HashAlgorithm)
	if o.Hasher == nil {
		return nil, errors.ErrUnknown(compdesc.KIND_HASH_ALGORITHM, digest.HashAlgorithm)
	}
	return &opts, nil
}
