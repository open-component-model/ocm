// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package signing

import (
	"crypto/x509"
	"strings"

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
	n := strings.TrimSpace(o.name)
	if n != "" {
		opts.SignatureNames = append(append([]string{}, n), opts.SignatureNames...)
	}
	opts.Signer = o.signer
}

////////////////////////////////////////////////////////////////////////////////

type verifier struct {
	name string
}

func VerifySignature(names ...string) Option {
	name := ""
	for _, n := range names {
		n = strings.TrimSpace(n)
		if n != "" {
			name = n
			break
		}
	}
	return &verifier{name}
}

func (o *verifier) ApplySigningOption(opts *Options) {
	opts.VerifySignature = true
	if o.name != "" {
		opts.SignatureNames = append(opts.SignatureNames, o.name)
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

type issuer struct {
	name string
}

func Issuer(name string) Option {
	return &issuer{name}
}

func (o *issuer) ApplySigningOption(opts *Options) {
	opts.Issuer = o.name
}

////////////////////////////////////////////////////////////////////////////////

type rootverts struct {
	pool *x509.CertPool
}

func RootCertificates(pool *x509.CertPool) Option {
	return &rootverts{pool}
}

func (o *rootverts) ApplySigningOption(opts *Options) {
	opts.RootCerts = o.pool
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
	Update            bool
	Recursively       bool
	Verify            bool
	Signer            signing.Signer
	Issuer            string
	VerifySignature   bool
	RootCerts         *x509.CertPool
	Hasher            signing.Hasher
	Keys              signing.KeyRegistry
	Registry          signing.Registry
	Resolver          ocm.ComponentVersionResolver
	SkipAccessTypes   map[string]bool
	SignatureNames    []string
	NormalizationAlgo string
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
	if o.VerifySignature {
		opts.VerifySignature = o.VerifySignature
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
	if len(o.SignatureNames) != 0 {
		opts.SignatureNames = o.SignatureNames
	}
	if o.SkipAccessTypes != nil {
		if opts.SkipAccessTypes == nil {
			opts.SkipAccessTypes = map[string]bool{}
		}
		for k, v := range o.SkipAccessTypes {
			opts.SkipAccessTypes[k] = v
		}
	}
	if o.Issuer != "" {
		opts.Issuer = o.Issuer
	}
	opts.Recursively = o.Recursively
	opts.Update = o.Update
	opts.Verify = o.Verify
	if o.NormalizationAlgo != "" {
		opts.NormalizationAlgo = o.NormalizationAlgo
	}
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
		if len(o.SignatureNames) == 0 {
			return errors.Newf("signature name required for signing")
		}
		if o.PrivateKey() == nil {
			return errors.ErrNotFound(compdesc.KIND_PRIVATE_KEY, o.SignatureNames[0])
		}
	}
	if o.VerifySignature {
		if len(o.SignatureNames) > 0 {
			for _, n := range o.SignatureNames {
				if pub := o.PublicKey(n); pub == nil {
					return errors.ErrNotFound(compdesc.KIND_PUBLIC_KEY, n)
				} else {
					err := o.checkCert(pub, n)
					if err != nil {
						return err
					}
				}
			}
		}
	} else {
		if o.Signer != nil {
			if pub := o.PublicKey(o.SignatureName()); pub != nil {
				o.VerifySignature = true
				err := o.checkCert(pub, o.SignatureName())
				if err != nil {
					return err
				}
			}
		}
	}
	if o.VerifySignature || o.Signer != nil {
		if o.NormalizationAlgo == "" {
			o.NormalizationAlgo = compdesc.JsonNormalisationV1
		}
	}
	if o.Hasher == nil {
		o.Hasher = o.Registry.GetHasher(sha256.Algorithm)
	}
	return nil
}

func (o *Options) checkCert(data interface{}, name string) error {
	cert, err := signing.GetCertificate(data)
	if err != nil {
		return nil
	}
	err = signing.VerifyCert(nil, o.RootCerts, "", cert)
	if err != nil {
		return errors.Wrapf(err, "public key %q", name)
	}
	return nil
}

func (o *Options) DoUpdate() bool {
	return o.Update || o.Signer != nil
}

func (o *Options) DoSign() bool {
	return o.Signer != nil && len(o.SignatureNames) > 0
}

func (o *Options) DoVerify() bool {
	return o.VerifySignature
}

func (o *Options) SignatureName() string {
	if len(o.SignatureNames) > 0 {
		return o.SignatureNames[0]
	}
	return ""
}

func (o *Options) SignatureConfigured(name string) bool {
	for _, n := range o.SignatureNames {
		if n == name {
			return true
		}
	}
	return false
}

func (o *Options) PublicKey(sig string) interface{} {
	if o.Keys != nil {
		k := o.Keys.GetPublicKey(sig)
		if k != nil {
			return k
		}
	}
	return o.Registry.GetPublicKey(sig)
}

func (o *Options) PrivateKey() interface{} {
	if o.Keys != nil {
		k := o.Keys.GetPrivateKey(o.SignatureName())
		if k != nil {
			return k
		}
	}
	return o.Registry.GetPrivateKey(o.SignatureName())
}

func (o *Options) For(digest *metav1.DigestSpec) (*Options, error) {
	opts := *o
	opts.VerifySignature = false // TODO: may be we want a mode to verify signature if present
	if !opts.Recursively {
		opts.Signer = nil
		opts.Update = false
	}
	if digest != nil {
		opts.Hasher = opts.Registry.GetHasher(digest.HashAlgorithm)
		if opts.Hasher == nil {
			return nil, errors.ErrUnknown(compdesc.KIND_HASH_ALGORITHM, digest.HashAlgorithm)
		}
	}
	return &opts, nil
}
