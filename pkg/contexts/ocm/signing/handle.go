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
	"context"
	"fmt"
	"reflect"

	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/printer"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/errors"
)

func ToDigestSpec(v interface{}) *metav1.DigestSpec {
	if v == nil {
		return nil
	}
	return v.(*metav1.DigestSpec)
}

func Apply(ctx context.Context, state *common.WalkingState, cv ocm.ComponentVersionAccess, opts *Options) (*metav1.DigestSpec, error) {
	if state == nil {
		s := common.NewWalkingState()
		state = &s
	}
	return apply(ctx, *state, cv, opts)
}

func apply(ctx context.Context, state common.WalkingState, cv ocm.ComponentVersionAccess, opts *Options) (*metav1.DigestSpec, error) {
	nv := common.VersionedElementKey(cv)
	if ok, err := state.Add(ocm.KIND_COMPONENTVERSION, nv); !ok {
		return ToDigestSpec(state.Closure[nv]), err
	}

	cd := cv.GetDescriptor().Copy()
	printer.Printf(ctx, "applying to version %q...\n", nv)
	for i, reference := range cd.References {
		var calculatedDigest *metav1.DigestSpec
		if reference.Digest == nil && !opts.DoUpdate() {
			return nil, errors.Newf(refMsg(reference, state, "no digest for component reference"))
		}
		if reference.Digest == nil || opts.Verify {
			nested, err := opts.Resolver.LookupComponentVersion(reference.GetComponentName(), reference.GetVersion())
			if err != nil {
				return nil, errors.Wrapf(err, refMsg(reference, state, "failed resolving component reference"))
			}
			closer := accessio.OnceCloser(nested)
			defer closer.Close()
			opts, err := opts.For(reference.Digest)
			if err != nil {
				return nil, errors.Wrapf(err, refMsg(reference, state, "failed resolving hasher for existing digest for component reference"))
			}
			calculatedDigest, err = apply(printer.WithGap(ctx, "  "), state, nested, opts)
			if err != nil {
				return nil, errors.Wrapf(err, refMsg(reference, state, "failed applying to component reference"))
			}
			closer.Close()
		}

		if reference.Digest == nil {
			cd.References[i].Digest = calculatedDigest
		} else {
			if calculatedDigest != nil && !reflect.DeepEqual(reference.Digest, calculatedDigest) {
				return nil, errors.Newf(refMsg(reference, state, "calculated reference digest (%+v) mismatches existing digest (%+v) for", calculatedDigest, reference.Digest))
			}
		}
		printer.Printf(ctx,"  reference %d:  %s:%s: digest %s\n", i, reference.ComponentName, reference.Version, calculatedDigest)
	}

	blobdigesters := cv.GetContext().BlobDigesters()
	for i, res := range cv.GetResources() {
		raw := &cd.Resources[i]
		acc, err := res.Access()
		if err != nil {
			return nil, errors.Wrapf(err, resMsg(raw, state, "failed getting access for resource"))
		}
		if _, ok := opts.SkipAccessTypes[acc.GetKind()]; ok {
			// set the do not sign digest notation on skip-access-type resources
			cd.Resources[i].Digest = metav1.NewExcludeFromSignatureDigest()
			continue
		}
		// special digest notation indicates to not digest the content
		if cd.Resources[i].Digest != nil && reflect.DeepEqual(cd.Resources[i].Digest, metav1.NewExcludeFromSignatureDigest()) {
			continue
		}

		meth, err := acc.AccessMethod(cv)
		if err != nil {
			return nil, errors.Wrapf(err, resMsg(raw, state, "failed creating access for resource"))
		}
		var req []cpi.DigesterType
		if raw.Digest != nil {
			req = []cpi.DigesterType{
				cpi.DigesterType{
					HashAlgorithm:          raw.Digest.HashAlgorithm,
					NormalizationAlgorithm: raw.Digest.NormalisationAlgorithm,
				},
			}
		}
		digest, err := blobdigesters.DetermineDigests(res.Meta().GetType(), opts.Hasher, opts.Registry, meth, req...)
		if err != nil {
			return nil, errors.Wrapf(err, resMsg(raw, state, "failed determining digest for resource"))
		}
		if len(digest) == 0 {
			return nil, errors.Newf(resMsg(raw, state, "no digester accepts resource"))
		}
		if raw.Digest != nil && !reflect.DeepEqual(*raw.Digest, digest[0]) {
			return nil, errors.Newf(resMsg(raw, state, "calculated resource digest (%+v) mismatches existing digest (%+v) for", digest, raw.Digest))
		}
		cd.Resources[i].Digest = &digest[0]
		printer.Printf(ctx, "  resource %d:  %s: digest %s\n", i, res.Meta().GetIdentity(cv.GetDescriptor().Resources), &digest[0])
	}
	digest, err := compdesc.Hash(cd, opts.NormalizationAlgo, opts.Hasher.Create())
	if err != nil {
		return nil, errors.Wrapf(err, "failed hashing component descriptor %s ", state.History)
	}
	spec := &metav1.DigestSpec{
		HashAlgorithm:          opts.Hasher.Algorithm(),
		NormalisationAlgorithm: compdesc.JsonNormalisationV1,
		Value:                  digest,
	}

	if opts.DoVerify() {
		list := opts.SignatureNames
		if len(opts.SignatureNames) == 0 {
			for _, s := range cd.Signatures {
				list = append(list, s.Name)
			}
			if len(list) == 0 {
				return nil, errors.Newf("no signature found in %s", state.History)
			}
		}
		found := []string{}
		for _, n := range list {
			f := cd.GetSignatureIndex(n)
			if f < 0 {
				continue
			}
			pub := opts.PublicKey(n)
			if pub == nil {
				if opts.SignatureConfigured(n) {
					return nil, errors.ErrNotFound(compdesc.KIND_PUBLIC_KEY, n, state.History.String())
				}
				printer.Printf(ctx, "Warning: no public key for signature %q in %s\n", n, state.History)
				continue
			}
			sig := &cd.Signatures[f]
			verifier := opts.Registry.GetVerifier(sig.Signature.Algorithm)
			if verifier == nil {
				if opts.SignatureConfigured(n) {
					return nil, errors.ErrUnknown(compdesc.KIND_VERIFY_ALGORITHM, n, state.History.String())
				}
				printer.Printf(ctx, "Warning: no verifier (%s) found for signature %q in %s\n", sig.Signature.Algorithm, n, state.History)
				continue
			}
			hasher := opts.Registry.GetHasher(sig.Digest.HashAlgorithm)
			if hasher == nil {
				return nil, errors.ErrUnknown(compdesc.KIND_HASH_ALGORITHM, sig.Digest.HashAlgorithm, state.History.String())
			}
			err = verifier.Verify(sig.Digest.Value, hasher.Crypto(), sig.ConvertToSigning(), pub)
			if err != nil {
				return nil, errors.ErrInvalidWrap(err, compdesc.KIND_SIGNATURE, sig.Signature.Algorithm, state.History.String())
			}
			found = append(found, n)
		}
		if len(found) == 0 {
			if !opts.DoSign() {
				return nil, errors.Newf("no verifiable signature found in %s", state.History)
			}
		}
	}
	found := cd.GetSignatureIndex(opts.SignatureName())
	if opts.DoSign() && (!opts.DoVerify() || found == -1) {
		sig, err := opts.Signer.Sign(digest, opts.Hasher.Crypto(), opts.Issuer, opts.PrivateKey())
		if err != nil {
			return nil, errors.Wrapf(err, "failed signing component descriptor %s ", state.History)
		}
		if sig.Issuer != "" {
			if opts.Issuer != "" && opts.Issuer != sig.Issuer {
				return nil, errors.Newf("signature issuer %q does not match intended issuer %q in %s", sig.Issuer, opts.Issuer, state.History)
			}
		} else {
			sig.Issuer = opts.Issuer
		}
		signature := metav1.Signature{
			Name:   opts.SignatureName(),
			Digest: *spec,
			Signature: metav1.SignatureSpec{
				Algorithm: sig.Algorithm,
				Value:     sig.Value,
				MediaType: sig.MediaType,
				Issuer:    sig.Issuer,
			},
		}
		if found >= 0 {
			cd.Signatures[found] = signature
		} else {
			cd.Signatures = append(cd.Signatures, signature)
		}
	}
	if opts.DoUpdate() {
		orig := cv.GetDescriptor()
		for i, res := range cd.Resources {
			orig.Resources[i].Digest = res.Digest
		}
		for i, res := range cd.References {
			orig.References[i].Digest = res.Digest
		}
		if opts.DoSign() {
			orig.Signatures = cd.Signatures
		}
	}
	state.Closure[nv] = spec
	return spec, nil
}

func refMsg(ref compdesc.ComponentReference, state common.WalkingState, msg string, args ...interface{}) string {
	return fmt.Sprintf("%s %q [%s:%s] in %s", fmt.Sprintf(msg, args...), ref.Name, ref.ComponentName, ref.Version, state.History)
}

func resMsg(ref *compdesc.Resource, state common.WalkingState, msg string, args ...interface{}) string {
	return fmt.Sprintf("%s %s:%s in %s", fmt.Sprintf(msg, args...), ref.Name, ref.Version, state.History)
}
