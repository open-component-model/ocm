package signing

import (
	"fmt"
	"io"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/general"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/cpi/accspeccpi"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/none"
	"ocm.software/ocm/api/ocm/extensions/attrs/signingattr"
	common "ocm.software/ocm/api/utils/misc"
)

// VerifyResourceDigestByResourceAccess verifies the digest of a resource passed by ResourceAccess.
func VerifyResourceDigestByResourceAccess(cv ocm.ComponentVersionAccess, rAcc ocm.ResourceAccess, bacc ocm.DataAccess, ostore ...VerifiedStore) (bool, error) {
	cd := cv.GetDescriptor()

	index := cd.GetResourceIndex(rAcc.Meta())
	if index < 0 {
		return false, errors.ErrNotFound("resource")
	}

	return VerifyResourceDigest(cv, index, bacc, ostore...)
}

// VerifyResourceDigest verify the digest of a resource taken from a component version.
// The data of the resources (typically after fetching the content) is given by a ocm.DataAccess.
// The digest info for verification is taken from the resource described by a component version,
// which has been used to retrieve the data. The component version itself is not verified.
//
// If a VerifiedStore is given, the found resource digest must match the content
// described by the store (which should describe verified component version descriptors),
// otherwise the verification is rejected. If the store does not contain the verified component version,
// verification is rejected, because no verified resource digest can be determined.
// So, if no store is give, the given component version is assumed to be already verified.
// Only the digest of the downloaded resource is verified.
//
// The function returns true if the verification has been executed. If an error occurs, or
// the verification has been failed, an appropriate error occurs.
// If the resource is not signature relevant (false,nil) is returned.
func VerifyResourceDigest(cv ocm.ComponentVersionAccess, i int, bacc ocm.DataAccess, ostore ...VerifiedStore) (bool, error) {
	octx := cv.GetContext()
	cd := cv.GetDescriptor()
	raw := &cd.Resources[i]

	// Check if the resource is signature relevant
	acc, err := octx.AccessSpecForSpec(raw.Access)
	if err != nil {
		return false, errors.Wrapf(err, resMsg(raw, "", "failed getting access for resource"))
	}

	if none.IsNone(acc.GetKind()) {
		return false, nil
	}
	if raw.Digest == nil {
		return false, errors.ErrNotFound("digest")
	}
	// special digest notation indicates to not digest the content
	if raw.Digest.IsExcluded() {
		return false, nil
	}

	// Check if the resource has already been verified
	store := general.Optional(ostore...)
	if store != nil {
		vcd := store.Get(cv)
		if vcd == nil {
			return false, fmt.Errorf("component version %s not verified", common.VersionedElementKey(cv))
		}
		if !vcd.Resources[i].Digest.Equal(raw.Digest) {
			return false, fmt.Errorf("component version %s corrupted", common.VersionedElementKey(cv))
		}
	}

	meth, err := acc.AccessMethod(cv)
	if err != nil {
		return false, errors.Wrapf(err, resMsg(raw, acc.Describe(octx), "failed creating access for resource"))
	}
	defer meth.Close()

	meth = NewRedirectedAccessMethod(meth, bacc)
	rdigest := raw.Digest

	dtype := DigesterType(rdigest)
	req := []ocm.DigesterType{dtype}

	registry := signingattr.Get(octx).HandlerRegistry()
	hasher := registry.GetHasher(dtype.HashAlgorithm)
	digest, err := octx.BlobDigesters().DetermineDigests(raw.Type, hasher, registry, meth, req...)
	if err != nil {
		return false, errors.Wrap(err, resMsg(raw, acc.Describe(octx), "failed determining digest for resource"))
	}
	if len(digest) == 0 {
		return false, errors.New(resMsg(raw, acc.Describe(octx), "no digester accepts resource"))
	}
	if !checkDigest(rdigest, &digest[0]) {
		return true, errors.New(resMsg(raw, acc.Describe(octx), "calculated resource digest (%+v) mismatches existing digest (%+v) for", &digest[0], rdigest))
	}
	return true, nil
}

type redirectedAccessMethod struct {
	ocm.AccessMethod
	acc ocm.DataAccess
}

var _ accspeccpi.AccessMethodView = (*redirectedAccessMethod)(nil)

func NewRedirectedAccessMethod(m ocm.AccessMethod, bacc ocm.DataAccess) ocm.AccessMethod {
	return &redirectedAccessMethod{m, bacc}
}

func (m *redirectedAccessMethod) Unwrap() interface{} {
	v, ok := m.AccessMethod.(accspeccpi.AccessMethodView)
	if !ok {
		return nil
	}

	return v.Unwrap()
}

func (m *redirectedAccessMethod) Close() error {
	list := errors.ErrList()
	list.Add(m.acc.Close())
	list.Add(m.AccessMethod.Close())
	return list.Result()
}

func (m *redirectedAccessMethod) Reader() (io.ReadCloser, error) {
	return m.acc.Reader()
}

func (m *redirectedAccessMethod) Get() ([]byte, error) {
	return m.acc.Get()
}
