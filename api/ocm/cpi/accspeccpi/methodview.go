package accspeccpi

import (
	"io"

	"github.com/modern-go/reflect2"
	"github.com/opencontainers/go-digest"

	"ocm.software/ocm/api/credentials"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/utils"
	"ocm.software/ocm/api/utils/blobaccess/blobaccess"
	"ocm.software/ocm/api/utils/refmgmt"
)

type DigestSource interface {
	GetDigest() (digest.Digest, error)
}

// AccessMethodView provides access
// to the implementation object behind an
// access method.
type AccessMethodView interface {
	utils.Unwrappable
	AccessMethod
}

// AccessMethodForImplementation wrap an access method implementation object
// into a published multi-view AccessMethod object. The original method implementation is
// closed when the last view is closed.
func AccessMethodForImplementation(acc AccessMethodImpl, err error) (AccessMethod, error) {
	if err != nil {
		if !reflect2.IsNil(acc) {
			acc.Close()
		}
		return nil, err
	}
	return refmgmt.WithView[AccessMethodImpl, AccessMethod](acc, accessMethodViewCreator), err
}

// BlobAccessForAccessSpec provide a blob access for an access specification.
func BlobAccessForAccessSpec(spec AccessSpec, cv ComponentVersionAccess) (blobaccess.BlobAccess, error) {
	m, err := spec.AccessMethod(cv)
	if err != nil {
		return nil, err
	}
	return m.AsBlobAccess(), nil
}

func accessMethodViewCreator(impl AccessMethodImpl, view *refmgmt.View[AccessMethod]) AccessMethod {
	if _, ok := impl.(DigestSpecProvider); ok {
		return &accessMethodViewWithDigest{accessMethodView{view, impl}}
	}
	return &accessMethodView{view, impl}
}

type accessMethodView struct {
	*refmgmt.View[AccessMethod]
	methodimpl AccessMethodImpl
}

type accessMethodViewWithDigest struct {
	accessMethodView
}

var (
	_ AccessMethodView                     = (*accessMethodView)(nil)
	_ credentials.ConsumerIdentityProvider = (*accessMethodView)(nil)
	_ DigestSpecProvider                   = (*accessMethodViewWithDigest)(nil)
)

func (a *accessMethodViewWithDigest) GetDigestSpec() (*metav1.DigestSpec, error) {
	return a.methodimpl.(DigestSpecProvider).GetDigestSpec()
}

func (a *accessMethodView) Unwrap() interface{} {
	return a.methodimpl
}

func (a *accessMethodView) AsBlobAccess() blobaccess.BlobAccess {
	return blobaccess.ForDataAccess("", -1, a.MimeType(), a)
}

func (a *accessMethodView) IsLocal() bool {
	return a.methodimpl.IsLocal()
}

func (a *accessMethodView) Get() ([]byte, error) {
	var result []byte
	err := a.Execute(func() (err error) {
		result, err = a.methodimpl.Get()
		return
	})
	return result, err
}

func (a *accessMethodView) Reader() (io.ReadCloser, error) {
	var result io.ReadCloser
	err := a.Execute(func() (err error) {
		result, err = a.methodimpl.Reader()
		return
	})
	return result, err
}

func (a *accessMethodView) GetKind() string {
	return a.methodimpl.GetKind()
}

func (a *accessMethodView) AccessSpec() AccessSpec {
	return a.methodimpl.AccessSpec()
}

func (a *accessMethodView) MimeType() string {
	return a.methodimpl.MimeType()
}

func (a *accessMethodView) GetConsumerId(uctx ...credentials.UsageContext) credentials.ConsumerIdentity {
	if p, ok := a.methodimpl.(credentials.ConsumerIdentityProvider); ok {
		return p.GetConsumerId(uctx...)
	}
	return nil
}

func (a *accessMethodView) GetIdentityMatcher() string {
	if p, ok := a.methodimpl.(credentials.ConsumerIdentityProvider); ok {
		return p.GetIdentityMatcher()
	}
	return ""
}

////////////////////////////////////////////////////////////////////////////////

func BlobAccessForAccessMethod(m AccessMethod) (blobaccess.AnnotatedBlobAccess[AccessMethodView], error) {
	m, err := m.Dup()
	if err != nil {
		return nil, err
	}
	return blobaccess.ForDataAccess("", -1, m.MimeType(), m.(AccessMethodView)), nil
}

func GetAccessMethodImplementation(m AccessMethod) interface{} {
	if v, ok := m.(AccessMethodView); ok {
		return v.Unwrap()
	}
	return nil
}
