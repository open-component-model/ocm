package accspeccpi

import (
	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/ocm/internal"
)

type (
	Context         = internal.Context
	ContextProvider = internal.ContextProvider

	AccessType = internal.AccessType

	AccessMethodImpl      = internal.AccessMethodImpl
	AccessMethod          = internal.AccessMethod
	UniformAccessSpecInfo = internal.UniformAccessSpecInfo
	AccessSpec            = internal.AccessSpec
	AccessSpecRef         = internal.AccessSpecRef

	HintProvider            = internal.HintProvider
	GlobalAccessProvider    = internal.GlobalAccessProvider
	CosumerIdentityProvider = credentials.ConsumerIdentityProvider

	ComponentVersionAccess = internal.ComponentVersionAccess
	DigestSpecProvider     = internal.DigestSpecProvider
)

var (
	newStrictAccessTypeScheme = internal.NewStrictAccessTypeScheme
	defaultAccessTypeScheme   = internal.DefaultAccessTypeScheme
)

func NewAccessSpecRef(spec AccessSpec) *AccessSpecRef {
	return internal.NewAccessSpecRef(spec)
}
