// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

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
	"github.com/open-component-model/ocm/pkg/utils"
)

type VersionInfo struct {
	Descriptor *compdesc.ComponentDescriptor
	Digest     *metav1.DigestSpec
}

func ToDigestSpec(v interface{}) *metav1.DigestSpec {
	if v == nil {
		return nil
	}
	return v.(*VersionInfo).Digest
}

type WalkingState = common.WalkingState[*VersionInfo]

func NewWalkingState() WalkingState {
	return common.NewWalkingState[*VersionInfo]()
}

func Apply(printer common.Printer, state *WalkingState, cv ocm.ComponentVersionAccess, opts *Options, closecv ...bool) (*metav1.DigestSpec, error) {
	if printer == nil {
		printer = common.NewPrinter(nil)
	}
	if state == nil {
		s := common.NewWalkingState[*VersionInfo]()
		state = &s
	}
	return apply(printer, *state, cv, opts, utils.Optional(closecv...))
}

func apply(printer common.Printer, state WalkingState, cv ocm.ComponentVersionAccess, opts *Options, closecv bool) (d *metav1.DigestSpec, efferr error) {
	var closer errors.ErrorFunction
	if closecv {
		closer = cv.Close
	}
	nv := common.VersionedElementKey(cv)
	defer errors.PropagateErrorf(&efferr, closer, "%s", state.History.Append(nv))

	if ok, err := state.Add(ocm.KIND_COMPONENTVERSION, nv); !ok {
		return ToDigestSpec(state.Closure[nv]), err
	}
	return _apply(printer, state, nv, cv, opts)
}

func _apply(printer common.Printer, state WalkingState, nv common.NameVersion, cv ocm.ComponentVersionAccess, opts *Options) (*metav1.DigestSpec, error) {
	cd := cv.GetDescriptor().Copy()
	octx := cv.GetContext()
	printer.Printf("applying to version %q...\n", nv)

	signatureNames := opts.SignatureNames
	if len(signatureNames) == 0 {
		for _, s := range cd.Signatures {
			signatureNames = append(signatureNames, s.Name)
		}
		if len(signatureNames) == 0 && opts.DoVerify() {
			return nil, errors.Newf("no signature found")
		}
	}
	if opts.DoVerify() && !opts.DoSign() {
		for _, n := range signatureNames {
			f := cd.GetSignatureIndex(n)
			if f < 0 {
				return nil, errors.Newf("signature %q not found", n)
			}
		}
	}

	if err := calculateReferenceDigests(printer, cd, state, opts); err != nil {
		return nil, err
	}

	blobdigesters := cv.GetContext().BlobDigesters()
	for i, res := range cv.GetResources() {
		raw := &cd.Resources[i]
		acc, err := res.Access()
		if err != nil {
			return nil, errors.Wrapf(err, resMsg(raw, "", "failed getting access for resource"))
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
			return nil, errors.Wrapf(err, resMsg(raw, acc.Describe(octx), "failed creating access for resource"))
		}
		var req []cpi.DigesterType
		if raw.Digest != nil {
			req = []cpi.DigesterType{
				{
					HashAlgorithm:          raw.Digest.HashAlgorithm,
					NormalizationAlgorithm: raw.Digest.NormalisationAlgorithm,
				},
			}
		}
		digest, err := blobdigesters.DetermineDigests(res.Meta().GetType(), opts.Hasher, opts.Registry, meth, req...)
		if err != nil {
			return nil, errors.Wrapf(err, resMsg(raw, acc.Describe(octx), "failed determining digest for resource"))
		}
		if len(digest) == 0 {
			return nil, errors.Newf(resMsg(raw, acc.Describe(octx), "no digester accepts resource"))
		}
		if raw.Digest != nil && !reflect.DeepEqual(*raw.Digest, digest[0]) {
			return nil, errors.Newf(resMsg(raw, acc.Describe(octx), "calculated resource digest (%+v) mismatches existing digest (%+v) for", digest, raw.Digest))
		}
		cd.Resources[i].Digest = &digest[0]
		printer.Printf("  resource %d:  %s: digest %s\n", i, res.Meta().GetIdentity(cv.GetDescriptor().Resources), &digest[0])
	}
	digest, err := compdesc.Hash(cd, opts.NormalizationAlgo, opts.Hasher.Create())
	if err != nil {
		return nil, errors.Wrapf(err, "failed hashing component descriptor")
	}
	spec := &metav1.DigestSpec{
		HashAlgorithm:          opts.Hasher.Algorithm(),
		NormalisationAlgorithm: opts.NormalizationAlgo,
		Value:                  digest,
	}

	if opts.DoVerify() {
		if err := doVerify(printer, cd, state, signatureNames, opts); err != nil {
			return nil, err
		}
	}

	found := cd.GetSignatureIndex(opts.SignatureName())
	if opts.DoSign() && (!opts.DoVerify() || found == -1) {
		sig, err := opts.Signer.Sign(digest, opts.Hasher.Crypto(), opts.Issuer, opts.PrivateKey())
		if err != nil {
			return nil, errors.Wrapf(err, "failed signing component descriptor")
		}
		if sig.Issuer != "" {
			if opts.Issuer != "" && opts.Issuer != sig.Issuer {
				return nil, errors.Newf("signature issuer %q does not match intended issuer %q", sig.Issuer, opts.Issuer)
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
	state.Closure[nv] = &VersionInfo{
		Descriptor: cd,
		Digest:     spec,
	}
	return spec, nil
}

func refMsg(ref compdesc.ComponentReference, msg string, args ...interface{}) string {
	return fmt.Sprintf("%s %s", fmt.Sprintf(msg, args...), ref)
}

func resMsg(ref *compdesc.Resource, acc string, msg string, args ...interface{}) string {
	if acc != "" {
		return fmt.Sprintf("%s %s:%s (%s)", fmt.Sprintf(msg, args...), ref.Name, ref.Version, acc)
	}
	return fmt.Sprintf("%s %s:%s", fmt.Sprintf(msg, args...), ref.Name, ref.Version)
}

func doVerify(printer common.Printer, cd *compdesc.ComponentDescriptor, state WalkingState, signatureNames []string, opts *Options) error {
	var err error
	found := []string{}
	for _, n := range signatureNames {
		f := cd.GetSignatureIndex(n)
		if f < 0 {
			continue
		}
		pub := opts.PublicKey(n)
		if pub == nil {
			if opts.SignatureConfigured(n) {
				return errors.ErrNotFound(compdesc.KIND_PUBLIC_KEY, n)
			}
			printer.Printf("Warning: no public key for signature %q in %s\n", n, state.History)
			continue
		}
		sig := &cd.Signatures[f]
		verifier := opts.Registry.GetVerifier(sig.Signature.Algorithm)
		if verifier == nil {
			if opts.SignatureConfigured(n) {
				return errors.ErrUnknown(compdesc.KIND_VERIFY_ALGORITHM, n)
			}
			printer.Printf("Warning: no verifier (%s) found for signature %q in %s\n", sig.Signature.Algorithm, n, state.History)
			continue
		}
		hasher := opts.Registry.GetHasher(sig.Digest.HashAlgorithm)
		if hasher == nil {
			return errors.ErrUnknown(compdesc.KIND_HASH_ALGORITHM, sig.Digest.HashAlgorithm)
		}
		err = verifier.Verify(sig.Digest.Value, hasher.Crypto(), sig.ConvertToSigning(), pub)
		if err != nil {
			return errors.ErrInvalidWrap(err, compdesc.KIND_SIGNATURE, sig.Signature.Algorithm)
		}
		found = append(found, n)
	}
	if len(found) == 0 {
		if !opts.DoSign() {
			return errors.Newf("no verifiable signature found")
		}
	}

	return nil
}

func calculateReferenceDigests(printer common.Printer, cd *compdesc.ComponentDescriptor, state WalkingState, opts *Options) error {
	for i, reference := range cd.References {
		var calculatedDigest *metav1.DigestSpec
		if reference.Digest == nil && !opts.DoUpdate() {
			printer.Printf("  no digest given for reference %s", reference)
		}
		if reference.Digest == nil || opts.Recursively || opts.Verify {
			nested, err := opts.Resolver.LookupComponentVersion(reference.GetComponentName(), reference.GetVersion())
			if err != nil {
				return errors.Wrapf(err, refMsg(reference, "failed resolving component reference"))
			}
			closer := accessio.OnceCloser(nested)
			defer closer.Close()
			digestOpts, err := opts.For(reference.Digest)
			if err != nil {
				return errors.Wrapf(err, refMsg(reference, "failed resolving hasher for existing digest for component reference"))
			}
			calculatedDigest, err = apply(printer.AddGap("  "), state, nested, digestOpts, true)
			if err != nil {
				return errors.Wrapf(err, refMsg(reference, "failed applying to component reference"))
			}
		} else {
			printer.Printf("  accepting digest from reference %s", reference)
			calculatedDigest = reference.Digest
		}

		if reference.Digest == nil {
			cd.References[i].Digest = calculatedDigest
		} else if calculatedDigest != nil && !reflect.DeepEqual(reference.Digest, calculatedDigest) {
			return errors.Newf(refMsg(reference, "calculated reference digest (%+v) mismatches existing digest (%+v) for", calculatedDigest, reference.Digest))
		}
		printer.Printf("  reference %d:  %s:%s: digest %s\n", i, reference.ComponentName, reference.Version, calculatedDigest)
	}
	return nil
}
