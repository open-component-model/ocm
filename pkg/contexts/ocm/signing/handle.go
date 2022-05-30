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
	"fmt"
	"reflect"

	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/common/accessio"
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

func Apply(printer common.Printer, cv ocm.ComponentVersionAccess, sign bool, opts *Options) (*metav1.DigestSpec, error) {
	if printer == nil {
		printer = common.NewPrinter(nil)
	}
	return apply(printer, common.NewWalkingState(), cv, sign, opts)
}

func apply(printer common.Printer, state common.WalkingState, cv ocm.ComponentVersionAccess, sign bool, opts *Options) (*metav1.DigestSpec, error) {
	nv := common.VersionedElementKey(cv)
	if ok, err := state.Add(ocm.KIND_COMPONENTVERSION, nv); !ok {
		return ToDigestSpec(state.Closure[nv]), err
	}

	cd := cv.GetDescriptor().Copy()
	printer.Printf("applying to version %q...\n", nv)
	for i, reference := range cd.ComponentReferences {
		nested, err := opts.Resolver.LookupComponentVersion(cd.GetName(), cd.GetVersion())
		if err != nil {
			return nil, errors.Wrapf(err, "failed resolving componentReference for %s:%s in %s", reference.Name, reference.Version, state.History)
		}
		closer := accessio.OnceCloser(nested)
		defer closer.Close()

		opts, err := opts.For(reference.Digest)
		if err != nil {
			return nil, errors.Wrapf(err, "failed resolving hasher for existing digest for %s:%s in %s", reference.Name, reference.Version, state.History)
		}
		digest, err := apply(printer, state, nested, sign && opts.Recursively, opts)
		if err != nil {
			return nil, errors.Wrapf(err, "failed applying to component version %s:%s: in", reference.Name, reference.Version, state.History)
		}
		if reference.Digest != nil && !reflect.DeepEqual(reference.Digest, digest) {
			return nil, fmt.Errorf("calculated cd reference digest mismatches existing digest %s:%s", reference.ComponentName, reference.Version)
		}
		closer.Close()
		cd.ComponentReferences[i].Digest = digest
	}

	blobdigesters := cv.GetContext().BlobDigesters()
	for i, res := range cv.GetResources() {
		acc, err := res.Access()

		if _, ok := opts.SkipAccessTypes[acc.GetKind()]; ok {
			// set the do not sign digest notation on skip-access-type resources
			cd.Resources[i].Digest = metav1.NewExcludeFromSignatureDigest()
			continue
		}
		// special digest notation indicates to not digest the content
		if cd.Resources[i].Digest != nil && reflect.DeepEqual(cd.Resources[i].Digest, metav1.NewExcludeFromSignatureDigest()) {
			continue
		}

		raw := &cd.Resources[i]
		meth, err := acc.AccessMethod(cv)
		if err != nil {
			return nil, errors.Wrapf(err, "failed creating access for resource for %s:%s in ", raw.Name, raw.Version, state.History)
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
			return nil, errors.Wrapf(err, "failed determining digest for resource %s:%s in ", raw.Name, raw.Version, state.History)
		}
		if len(digest) == 0 {
			return nil, errors.Newf("no digester accepts resource %s:%s in %s", raw.Name, raw.Version, state.History)
		}
		if raw.Digest != nil && !reflect.DeepEqual(raw.Digest, digest) {
			return nil, fmt.Errorf("calculated resource digest mismatches existing digest %s:%s in %s", raw.Name, raw.Version, state.History)
		}
		cd.Resources[i].Digest = &digest[0]
	}
	digest, err := compdesc.Hash(cd, compdesc.JsonNormalisationV1, opts.Hasher.Create())
	if err != nil {
		return nil, errors.Wrapf(err, "failed hashing component descriptor %s ", state.History)
	}
	spec := &metav1.DigestSpec{
		HashAlgorithm:          opts.Hasher.Algorithm(),
		NormalisationAlgorithm: compdesc.JsonNormalisationV1,
		Value:                  digest,
	}
	if sign {
		sig, media, err := opts.Signer.Sign(digest, opts.Registry.GetPrivateKey(opts.SignatureName))
		if err != nil {
			return nil, errors.Wrapf(err, "failed signing component descriptor %s ", state.History)
		}
		signature := metav1.Signature{
			Name:   opts.SignatureName,
			Digest: *spec,
			Signature: metav1.SignatureSpec{
				Algorithm: opts.Signer.Algorithm(),
				Value:     sig,
				MediaType: media,
			},
		}
		found := false
		for i, s := range cd.Signatures {
			if s.Name == opts.SignatureName {
				cd.Signatures[i] = signature
				found = true
				break
			}
		}
		if !found {
			cd.Signatures = append(cd.Signatures, signature)
		}
	}
	if opts.Update {
		orig := cv.GetDescriptor()
		for i, res := range cd.Resources {
			orig.Resources[i].Digest = res.Digest
		}
		if sign {
			orig.Signatures = cd.Signatures
		}
	}
	return spec, nil
}
