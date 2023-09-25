// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package signing

import (
	"fmt"
	"reflect"

	"github.com/mandelsoft/logging"

	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/none"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/signing"
	"github.com/open-component-model/ocm/pkg/utils"
)

var REALM = logging.NewRealm("signing")

func ArtefactDigest(r *compdesc.Resource) metav1.ArtefactDigest {
	return metav1.ArtefactDigest{
		Name:          r.Name,
		Version:       r.Version,
		ExtraIdentity: r.ExtraIdentity.Copy(),
		Digest:        *r.Digest,
	}
}

// VersionInfo keeps track of handled component versions
// and provides the digest context used for a dedicated root component
// this component version is digested for (by following component references).
type VersionInfo struct {
	digestingContexts map[common.NameVersion]*DigestContext
}

func NewVersionInfo(cd *compdesc.ComponentDescriptor, parent *DigestContext) (*VersionInfo, *DigestContext) {
	vi := &VersionInfo{
		digestingContexts: map[common.NameVersion]*DigestContext{},
	}
	return vi, vi.CreateContext(cd, parent)
}

func (vi *VersionInfo) GetContext(nv common.NameVersion) *DigestContext {
	return vi.digestingContexts[nv]
}

func (vi *VersionInfo) CreateContext(cd *compdesc.ComponentDescriptor, parent *DigestContext) *DigestContext {
	var key common.NameVersion
	if parent != nil {
		key = parent.CtxKey
	} else {
		key = common.VersionedElementKey(cd)
	}
	nctx := NewDigestContext(cd.Copy(), parent)

	// check for reuse of matching context
	if parent != nil {
		for _, ctx := range vi.digestingContexts {
			if ctx.ValidFor(nctx) {
				if err := nctx.Use(ctx); err != nil {
					panic(err)
				}
				nctx.Source = ctx.CtxKey
				break
			}
		}
	}
	if vi.digestingContexts[key] != nil {
		panic(fmt.Sprintf("duplicate creation of digest context %q for %q", nctx.Key, key))
	}
	vi.digestingContexts[key] = nctx
	return nctx
}

type WalkingState struct {
	common.WalkingState[*VersionInfo, *DigestContext]
}

func NewWalkingState(lctx ...logging.Context) WalkingState {
	return WalkingState{common.NewWalkingState[*VersionInfo, *DigestContext](nil, lctx...)}
}

func (s *WalkingState) GetContext(nv common.NameVersion, ctxkey common.NameVersion) *DigestContext {
	vi := s.Get(nv)
	if vi == nil {
		return nil
	}
	return vi.digestingContexts[ctxkey]
}

func Apply(printer common.Printer, state *WalkingState, cv ocm.ComponentVersionAccess, opts *Options, closecv ...bool) (*metav1.DigestSpec, error) {
	if printer != nil {
		opts = opts.Dup()
		opts.Printer = printer
	}
	err := opts.Complete(cv.GetContext())
	if err != nil {
		return nil, err
	}
	if state == nil {
		s := NewWalkingState(cv.GetContext().LoggingContext().WithContext(REALM))
		state = &s
	}
	dc, err := apply(*state, cv, opts, utils.Optional(closecv...))
	if err != nil {
		return nil, err
	}

	return dc.Digest, nil
}

func RequireReProcessing(vi *VersionInfo, ctx *DigestContext, opts *Options) bool {
	if vi == nil || ctx == nil || vi.digestingContexts[ctx.CtxKey] == nil {
		return true
	}
	return opts.DoSign() && !vi.digestingContexts[ctx.CtxKey].Signed
}

func apply(state WalkingState, cv ocm.ComponentVersionAccess, opts *Options, closecv bool) (dc *DigestContext, efferr error) {
	var closer errors.ErrorFunction
	if closecv {
		closer = func() error {
			return cv.Close()
		}
	}
	nv := common.VersionedElementKey(cv)
	defer errors.PropagateErrorf(&efferr, closer, "%s", state.History.Append(nv))

	vi := state.Get(nv)
	if ok, err := state.Add(ocm.KIND_COMPONENTVERSION, nv); !ok {
		if err != nil || !RequireReProcessing(vi, state.Context, opts) {
			return vi.digestingContexts[state.Context.CtxKey], err
		}
	}
	return _apply(state, nv, cv, vi, opts)
}

func _apply(state WalkingState, nv common.NameVersion, cv ocm.ComponentVersionAccess, vi *VersionInfo, opts *Options) (*DigestContext, error) { //nolint: maintidx // yes
	prefix := ""
	var ctx *DigestContext
	if vi == nil {
		vi, ctx = NewVersionInfo(cv.GetDescriptor(), state.Context)
	} else {
		prefix = "re"
		ctx = vi.CreateContext(cv.GetDescriptor(), state.Context)
	}

	if ctx.IsRoot() {
		ctx.RootContextInfo.Sign = opts.DoSign()
		// the first one creating hashes determines the digest mode to be used for all further signatures
		mode := GetDigestMode(cv.GetDescriptor(), opts.DigestMode)
		opts = opts.WithDigestMode(mode)
		if opts.DoSign() || !opts.DoVerify() {
			ctx.DigestType = ocm.DigesterType{
				HashAlgorithm:          opts.Hasher.Algorithm(),
				NormalizationAlgorithm: opts.NormalizationAlgo,
			}
		} else {
			var err error
			opts, err = ctx.determineSignatureInfo(state, opts)
			if err != nil {
				return nil, fmt.Errorf("failed to determine signature info: %w", err)
			}
		}
		ctx.Hasher = opts.Registry.GetHasher(ctx.DigestType.HashAlgorithm)
		if ctx.Hasher == nil {
			return nil, errors.ErrUnknown(compdesc.KIND_HASH_ALGORITHM, ctx.DigestType.NormalizationAlgorithm, "component version "+ctx.Key.String())
		}
	}
	if ctx.Digest != nil {
		if !opts.DoSign() || ctx.Signed {
			state.Logger.Debug("reusing from context", "cv", nv, "root", ctx.CtxKey, "ctx", ctx.Source)
			opts.Printer.Printf("reusing %s[%s] from context %q\n", nv, ctx.CtxKey, ctx.Source)
			return ctx, nil
		}
	}

	signed := false
	if ctx.Parent != nil && opts.DoSign() && GetDigestMode(cv.GetDescriptor(), opts.DigestMode) != opts.DigestMode {
		// recursive nested signing for an already somehow signed cv musts always be done
		// in actual digest context to use the already existing recursive resource
		// digests used for the existing signatures.
		substate := state
		substate.Context = nil
		nctx, err := _apply(substate, nv, cv, vi, opts)
		if err != nil {
			return nil, err
		}

		// check for contradiction in context. If there is no contradiction
		// the private context can be reused in actual context.
		if nctx.ValidFor(ctx) {
			ctx.Use(nctx)
			return nctx, nil
		}
		// after signing in own context, continue with verification in outer signing context
		opts = opts.StopRecursion()
		signed = true
	}

	cd := ctx.Descriptor
	state.Context = ctx

	state.Logger.Debug(fmt.Sprintf("%sapplying to version", prefix), "cv", nv, "root", ctx.CtxKey)
	opts.Printer.Printf("%sapplying to version %q[%s]...\n", prefix, nv, ctx.CtxKey)

	signatureNames := opts.SignatureNames
	if len(signatureNames) == 0 && opts.Keyless {
		return nil, errors.New("signature not provided")
	}
	if opts.DoVerify() && !opts.DoSign() {
		for _, n := range signatureNames {
			f := cd.GetSignatureIndex(n)
			if f < 0 {
				return nil, errors.Newf("signature %q not found", n)
			}
		}
	}

	var spec *metav1.DigestSpec
	legacy := signing.IsLegacyHashAlgorithm(ctx.RootContextInfo.DigestType.HashAlgorithm) && !opts.DoSign()
	if ctx.Digest == nil {
		if err := calculateReferenceDigests(state, opts, legacy); err != nil {
			return nil, err
		}
		if err := calculateResourceDigests(state, cv, cd, opts, legacy, ctx.GetPreset(ctx.Key)); err != nil {
			return nil, err
		}
		dt := ctx.DigestType
		if pre := ctx.GetPreset(ctx.Key); pre != nil && pre.Digest != nil {
			dt = DigesterType(pre.Digest)
		}
		hasher := opts.Registry.GetHasher(dt.HashAlgorithm)
		if hasher == nil {
			return nil, fmt.Errorf("unknown hash algorithm %q", dt.HashAlgorithm)
		}
		norm, digest, err := compdesc.NormHash(cd, dt.NormalizationAlgorithm, hasher.Create())
		if err != nil {
			return nil, errors.Wrapf(err, "failed hashing component descriptor")
		}
		state.Logger.Debug("component version digest", "cv", nv, "root", ctx.CtxKey, "digest", digest, "hashalgo", dt.HashAlgorithm, "normalgo", dt.NormalizationAlgorithm, "normalized", string(norm))
		spec = &metav1.DigestSpec{
			HashAlgorithm:          dt.HashAlgorithm,
			NormalisationAlgorithm: dt.NormalizationAlgorithm,
			Value:                  digest,
		}
	}

	if opts.DoVerify() {
		dig, err := doVerify(cd, state, signatureNames, opts)
		if err != nil {
			return nil, err
		}
		if dig != nil {
			spec = dig
		}
	}
	err := ctx.Propagate(spec)
	if err != nil {
		return nil, errors.Wrapf(err, "failed propagating digest context")
	}

	found := cd.GetSignatureIndex(opts.SignatureName())
	if opts.DoSign() && (!opts.DoVerify() || found == -1) {
		priv, err := opts.PrivateKey()
		if err != nil {
			return nil, err
		}
		sig, err := opts.Signer.Sign(cv.GetContext().CredentialsContext(), ctx.Digest.Value, opts.Hasher.Crypto(), opts.Issuer, priv)
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
			Digest: *ctx.Digest,
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
	state.Closure[nv] = vi

	if !signed && opts.DoUpdate() {
		orig := cv.GetDescriptor()
		state.Logger.Debug("updating digests", "cv", nv)
		for i, res := range cd.Resources {
			orig.Resources[i].Digest = res.Digest
		}
		if opts.StoreLocally() {
			for i, res := range cd.References {
				orig.References[i].Digest = res.Digest
			}
		} else {
			orig.NestedDigests = ctx.GetDigests()
		}
		if opts.DoSign() {
			state.Logger.Debug("update signature", "cv", nv)
			orig.Signatures = cd.Signatures
			ctx.Signed = true
		}
	}
	return ctx, nil
}

func checkDigest(orig *metav1.DigestSpec, act *metav1.DigestSpec) bool {
	if orig != nil {
		algo := signing.NormalizeHashAlgorithm(orig.HashAlgorithm)
		if algo == act.HashAlgorithm {
			act.HashAlgorithm = orig.HashAlgorithm
		}
		if !reflect.DeepEqual(orig, act) {
			return false
		}
	}
	return true
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

func doVerify(cd *compdesc.ComponentDescriptor, state WalkingState, signatureNames []string, opts *Options) (*metav1.DigestSpec, error) {
	var spec *metav1.DigestSpec
	found := []string{}
	for _, n := range signatureNames {
		f := cd.GetSignatureIndex(n)
		if f < 0 {
			continue
		}
		var pub any
		if !opts.Keyless {
			pub = opts.PublicKey(n)
			if pub == nil {
				if opts.SignatureConfigured(n) {
					return nil, errors.ErrNotFound(compdesc.KIND_PUBLIC_KEY, n)
				}
				opts.Printer.Printf("Warning: no public key for signature %q in %s\n", n, state.History)
				continue
			}
		}
		sig := &cd.Signatures[f]
		verifier := opts.Registry.GetVerifier(sig.Signature.Algorithm)
		if verifier == nil {
			if opts.SignatureConfigured(n) {
				return nil, errors.ErrUnknown(compdesc.KIND_VERIFY_ALGORITHM, n)
			}
			opts.Printer.Printf("Warning: no verifier (%s) found for signature %q in %s\n", sig.Signature.Algorithm, n, state.History)
			continue
		}

		hasher := opts.Registry.GetHasher(sig.Digest.HashAlgorithm)
		if hasher == nil {
			return nil, errors.ErrUnknown(compdesc.KIND_HASH_ALGORITHM, sig.Digest.HashAlgorithm)
		}
		digest, err := compdesc.Hash(cd, sig.Digest.NormalisationAlgorithm, hasher.Create())
		if err != nil {
			return nil, errors.Wrapf(err, "failed hashing component descriptor")
		}
		if sig.Digest.Value != digest {
			return nil, errors.Newf("signature digest (%s) does not match found digest (%s)", sig.Digest.Value, digest)
		}
		err = verifier.Verify(sig.Digest.Value, hasher.Crypto(), sig.ConvertToSigning(), pub)
		if err != nil {
			return nil, errors.ErrInvalidWrap(err, compdesc.KIND_SIGNATURE, sig.Signature.Algorithm)
		}
		found = append(found, n)
		if opts.SignatureName() == sig.Name {
			d := sig.Digest
			d.HashAlgorithm = signing.NormalizeHashAlgorithm(d.HashAlgorithm)
			spec = &d
		}
	}
	if len(found) == 0 {
		if !opts.DoSign() {
			return nil, errors.Newf("no verifiable signature found")
		}
	}

	return spec, nil
}

func calculateReferenceDigests(state WalkingState, opts *Options, legacy bool) error {
	ctx := state.Context
	cd := ctx.Descriptor
	for i, reference := range cd.References {
		rnv := ocm.ComponentRefKey(&reference)
		nctx := state.GetContext(rnv, state.Context.CtxKey)

		if nctx == nil || nctx.Digest == nil {
			opts.Printer.Printf("  no digest found for %q\n", rnv)
			nctx = nil
		}
		nested, err := opts.Resolver.LookupComponentVersion(reference.GetComponentName(), reference.GetVersion())
		if err != nil {
			return errors.Wrapf(err, refMsg(reference, "failed resolving component reference"))
		}
		if nctx == nil || opts.Recursively || opts.Verify {
			closer := accessio.OnceCloser(nested)
			defer closer.Close()
			digestOpts := opts.Nested()
			nctx, err = apply(state, nested, digestOpts, true)
			if err != nil {
				return errors.Wrapf(err, refMsg(reference, "failed applying to component reference"))
			}
		} else {
			state.Logger.Debug("accepting digest from context", "reference", reference)
			opts.Printer.Printf("  accepting digest from context for %s", reference)
			if err != nil {
				return errors.Wrapf(err, refMsg(reference, "failed applying to component reference"))
			}
		}
		if reference.Digest != nil {
			if ctx.IsRoot() {
				if DigesterType(reference.Digest) == DigesterType(nctx.Digest) {
					if nctx.Digest != nil && !reflect.DeepEqual(reference.Digest, nctx.Digest) {
						return errors.Newf(refMsg(reference, "calculated reference digest (%+v) mismatches existing digest (%+v) for", nctx.Digest, reference.Digest))
					}
				}
			}
			pre := ctx.In[nctx.Key]
			if pre != nil {
				if DigesterType(pre.Digest) == DigesterType(nctx.Digest) {
					if nctx.Digest != nil && !reflect.DeepEqual(pre.Digest, nctx.Digest) {
						return errors.Newf(refMsg(reference, "calculated reference digest (%+v) mismatches existing digest (%+v) for", nctx.Digest, reference.Digest))
					}
				}
			}
		}
		if legacy {
			nctx.Digest.HashAlgorithm = signing.LegacyHashAlgorithm(nctx.Digest.HashAlgorithm)
		}
		cd.References[i].Digest = nctx.Digest
		ctx.Refs[nctx.Key] = nctx.Digest
		state.Logger.Debug("reference digest", "index", i, "reference", common.NewNameVersion(reference.ComponentName, reference.Version), "hashalgo", nctx.Digest.HashAlgorithm, "normalgo", nctx.Digest.NormalisationAlgorithm, "digest", nctx.Digest.Value)
		opts.Printer.Printf("  reference %d:  %s:%s: digest %s\n", i, reference.ComponentName, reference.Version, nctx.Digest)
	}
	return nil
}

func calculateResourceDigests(state WalkingState, cv ocm.ComponentVersionAccess, cd *compdesc.ComponentDescriptor, opts *Options, legacy bool, preset *metav1.NestedComponentDigests) error {
	octx := cv.GetContext()
	blobdigesters := octx.BlobDigesters()
	for i, res := range cv.GetResources() {
		meta := res.Meta()
		preset := preset.Lookup(meta.Name, meta.Version, meta.ExtraIdentity)
		raw := &cd.Resources[i]
		acc, err := res.Access()
		if err != nil {
			return errors.Wrapf(err, resMsg(raw, "", "failed getting access for resource"))
		}
		if none.IsNone(acc.GetKind()) {
			cd.Resources[i].Digest = nil
			continue
		}
		if _, ok := opts.SkipAccessTypes[acc.GetKind()]; ok {
			// set the do not sign digest notation on skip-access-type resources
			// if no digest is already known.
			if cd.Resources[i].Digest == nil {
				cd.Resources[i].Digest = metav1.NewExcludeFromSignatureDigest()
			}
		}
		// special digest notation indicates to not digest the content
		if cd.Resources[i].Digest.IsExcluded() {
			continue
		}

		meth, err := acc.AccessMethod(cv)
		if err != nil {
			return errors.Wrapf(err, resMsg(raw, acc.Describe(octx), "failed creating access for resource"))
		}

		var rdigest *metav1.DigestSpec
		if raw.Digest != nil &&
			(state.Context.IsRoot() || opts.DigestMode != DIGESTMODE_TOP || raw.Digest.HashAlgorithm == opts.Hasher.Algorithm()) {
			// keep precalculated digest, if present.
			// For top mode any non-root level digest can be recalculated.
			rdigest = raw.Digest
		}
		if preset != nil && (!state.Context.RootContextInfo.Sign || preset.Digest.HashAlgorithm == opts.Hasher.Algorithm()) {
			// prefer digest from context.
			// If access method enforces a dedicated algorithm, then this should have been done
			// during the fist calculation, also, so, the same type should be used.
			rdigest = &preset.Digest
		}
		var req []ocm.DigesterType
		if rdigest != nil {
			req = []ocm.DigesterType{DigesterType(rdigest)}
		}
		digest, err := blobdigesters.DetermineDigests(res.Meta().GetType(), opts.Hasher, opts.Registry, meth, req...)
		if err != nil {
			return errors.Wrapf(err, resMsg(raw, acc.Describe(octx), "failed determining digest for resource"))
		}
		if len(digest) == 0 {
			return errors.Newf(resMsg(raw, acc.Describe(octx), "no digester accepts resource"))
		}
		if !checkDigest(rdigest, &digest[0]) {
			return errors.Newf(resMsg(raw, acc.Describe(octx), "calculated resource digest (%+v) mismatches existing digest (%+v) for", digest, rdigest))
		}
		if NormalizedDigesterType(raw.Digest) == NormalizedDigesterType(&digest[0]) {
			if !checkDigest(raw.Digest, &digest[0]) {
				return errors.Newf(resMsg(raw, acc.Describe(octx), "calculated resource digest (%+v) mismatches existing digest (%+v) for", digest, raw.Digest))
			}
		}
		cd.Resources[i].Digest = &digest[0]
		if legacy {
			cd.Resources[i].Digest.HashAlgorithm = signing.LegacyHashAlgorithm(cd.Resources[i].Digest.HashAlgorithm)
		}
		rid := res.Meta().GetIdentity(cv.GetDescriptor().Resources)
		state.Logger.Debug("resource digest", "index", i, "id", rid, "hashalgo", digest[0].HashAlgorithm, "normalgo", digest[0].NormalisationAlgorithm, "digest", digest[0].Value)
		opts.Printer.Printf("  resource %d:  %s: digest %s\n", i, rid, &digest[0])
	}
	return nil
}

func DigesterType(digest *metav1.DigestSpec) ocm.DigesterType {
	var dc ocm.DigesterType
	if digest != nil {
		dc.HashAlgorithm = digest.HashAlgorithm
		dc.NormalizationAlgorithm = digest.NormalisationAlgorithm
	}
	return dc
}

func NormalizedDigesterType(digest *metav1.DigestSpec) ocm.DigesterType {
	dc := DigesterType(digest)
	dc.HashAlgorithm = signing.NormalizeHashAlgorithm(dc.HashAlgorithm)
	return dc
}

// GetDigestMode checks whether the versio has already been digested.
// If so, the digest mode used at this time fixes the mode for all further
// signing processes.
// If a version is still undigested, any mode possible and is optionally
// defaulted by an additional argument.
func GetDigestMode(cd *compdesc.ComponentDescriptor, def ...string) string {
	if len(cd.NestedDigests) > 0 {
		return DIGESTMODE_TOP
	}
	if len(cd.References) > 0 {
		if cd.References[0].Digest != nil {
			return DIGESTMODE_LOCAL
		}
	}
	return utils.Optional(def...)
}
