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
	"github.com/open-component-model/ocm/pkg/signing/hasher/sha256"
)

type Option interface {
	ApplySigningOption(o *Options)
}

func evalFlag(flags ...bool) bool {
	flag := len(flags) == 0
	for _, f := range flags {
		flag = flag || f
	}
	return flag
}

////////////////////////////////////////////////////////////////////////////////

type recursive struct {
	flag bool
}

func Recursive(flags ...bool) Option {
	return &recursive{evalFlag(flags...)}
}

func (o *recursive) ApplySigningOption(opts *Options) {
	opts.Recursively = o.flag
}

////////////////////////////////////////////////////////////////////////////////

type update struct {
	flag bool
}

func Update(flags ...bool) Option {
	return &update{evalFlag(flags...)}
}

func (o *update) ApplySigningOption(opts *Options) {
	opts.Update = o.flag
}

////////////////////////////////////////////////////////////////////////////////

type verify struct {
	flag bool
}

func VerifyDigests(flags ...bool) Option {
	return &verify{evalFlag(flags...)}
}

func (o *verify) ApplySigningOption(opts *Options) {
	opts.Verify = o.flag
}

////////////////////////////////////////////////////////////////////////////////

type signer struct {
	signer signing.Signer
	name   string
}

func Sign(h signing.Signer, name string) Option {
	return &signer{h, name}
}

func (o *signer) ApplySigningOption(opts *Options) {
	opts.SignatureName = o.name
	opts.Signer = o.signer
}

////////////////////////////////////////////////////////////////////////////////

type verifier struct {
	name string
}

func VerifySignature(names ...string) Option {
	name := ""
	for _, n := range names {
		if n != "" {
			name = n
			break
		}
	}
	return &verifier{name}
}

func (o *verifier) ApplySigningOption(opts *Options) {
	opts.Verifier = true
	if o.name != "" {
		opts.SignatureName = o.name
	}
}

////////////////////////////////////////////////////////////////////////////////

type resolver struct {
	resolver []ocm.ComponentVersionResolver
}

func Resolver(h ...ocm.ComponentVersionResolver) Option {
	return &resolver{h}
}

func (o *resolver) ApplySigningOption(opts *Options) {
	opts.Resolver = ocm.NewCompoundResolver(append(append([]ocm.ComponentVersionResolver{}, opts.Resolver), o.resolver...)...)
}

////////////////////////////////////////////////////////////////////////////////

type skip struct {
	skip map[string]bool
}

func SkipAccessTypes(names ...string) Option {
	m := map[string]bool{}
	for _, n := range names {
		m[n] = true
	}
	return &skip{m}
}

func (o *skip) ApplySigningOption(opts *Options) {
	if len(o.skip) > 0 {
		if opts.SkipAccessTypes == nil {
			opts.SkipAccessTypes = map[string]bool{}
		}
		for k, v := range o.skip {
			opts.SkipAccessTypes[k] = v
		}
	}
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

type privkey struct {
	name string
	key  interface{}
}

func PrivateKey(name string, key interface{}) Option {
	return &privkey{name, key}
}

func (o *privkey) ApplySigningOption(opts *Options) {
	if opts.Keys == nil {
		opts.Keys = signing.NewKeyRegistry()
	}
	opts.Keys.RegisterPrivateKey(o.name, o.key)
}

////////////////////////////////////////////////////////////////////////////////

type pubkey struct {
	name string
	key  interface{}
}

func PublicKey(name string, key interface{}) Option {
	return &pubkey{name, key}
}

func (o *pubkey) ApplySigningOption(opts *Options) {
	if opts.Keys == nil {
		opts.Keys = signing.NewKeyRegistry()
	}
	opts.Keys.RegisterPublicKey(o.name, o.key)
}

////////////////////////////////////////////////////////////////////////////////

type Options struct {
	Update          bool
	Recursively     bool
	Verify          bool
	Signer          signing.Signer
	Verifier        bool
	Hasher          signing.Hasher
	Keys            signing.KeyRegistry
	Registry        signing.Registry
	Resolver        ocm.ComponentVersionResolver
	SkipAccessTypes map[string]bool
	SignatureName   string
}

var _ Option = (*Options)(nil)

func NewOptions(list ...Option) *Options {
	return (&Options{}).Eval(list...)
}

func (opts *Options) Eval(list ...Option) *Options {
	for _, o := range list {
		o.ApplySigningOption(opts)
	}
	return opts
}

func (o *Options) ApplySigningOption(opts *Options) {
	if o.Signer != nil {
		opts.Signer = o.Signer
	}
	if o.Verifier {
		opts.Verifier = o.Verifier
	}
	if o.Hasher != nil {
		opts.Hasher = o.Hasher
	}
	if o.Registry != nil {
		opts.Registry = o.Registry
	}
	if o.Resolver != nil {
		opts.Resolver = o.Resolver
	}
	if o.SignatureName != "" {
		opts.SignatureName = o.SignatureName
	}
	if o.SkipAccessTypes != nil {
		if opts.SkipAccessTypes == nil {
			opts.SkipAccessTypes = map[string]bool{}
		}
		for k, v := range o.SkipAccessTypes {
			opts.SkipAccessTypes[k] = v
		}
	}
	opts.Recursively = o.Recursively
	opts.Update = o.Update
	opts.Verify = o.Verify
}

func (o *Options) Complete(registry signing.Registry) error {
	if o.Registry == nil {
		if registry == nil {
			registry = signing.DefaultRegistry()
		}
		o.Registry = registry
	}
	if o.SkipAccessTypes == nil {
		o.SkipAccessTypes = map[string]bool{}
	}
	if o.Signer != nil {
		if o.SignatureName == "" {
			return errors.Newf("signature name required for signing")
		}
		if o.PrivateKey() == nil {
			return errors.ErrNotFound(compdesc.KIND_PRIVATE_KEY, o.SignatureName)
		}
	}
	if o.Verifier {
		if o.SignatureName == "" {
			return errors.Newf("signature name required for verifying signature")
		}
		if o.PublicKey() == nil {
			return errors.ErrNotFound(compdesc.KIND_PUBLIC_KEY, o.SignatureName)
		}
	} else {
		if o.Signer != nil {
			if o.PublicKey() != nil {
				o.Verifier = true
			}
		}
	}
	if o.Hasher == nil {
		o.Hasher = o.Registry.GetHasher(sha256.Algorithm)
	}
	return nil
}

func (o *Options) DoUpdate() bool {
	return o.Update || o.Signer != nil
}

func (o *Options) DoSign() bool {
	return o.Signer != nil && o.SignatureName != ""
}

func (o *Options) DoVerify() bool {
	return o.Verifier && o.SignatureName != ""
}

func (o *Options) PublicKey() interface{} {
	if o.Keys != nil {
		k := o.Keys.GetPublicKey(o.SignatureName)
		if k != nil {
			return k
		}
	}
	return o.Registry.GetPublicKey(o.SignatureName)
}

func (o *Options) PrivateKey() interface{} {
	if o.Keys != nil {
		k := o.Keys.GetPrivateKey(o.SignatureName)
		if k != nil {
			return k
		}
	}
	return o.Registry.GetPrivateKey(o.SignatureName)
}

func (o *Options) For(digest *metav1.DigestSpec) (*Options, error) {
	opts := *o
	if !opts.Recursively {
		opts.Signer = nil
		opts.Verifier = false
	}
	if digest != nil {
		opts.Hasher = opts.Registry.GetHasher(digest.HashAlgorithm)
		if opts.Hasher == nil {
			return nil, errors.ErrUnknown(compdesc.KIND_HASH_ALGORITHM, digest.HashAlgorithm)
		}
	}
	return &opts, nil
}
