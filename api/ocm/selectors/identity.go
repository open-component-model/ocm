package selectors

import (
	"regexp"

	"github.com/Masterminds/semver/v3"
	v1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/selectors/accessors"
)

type IdentitySelector interface {
	MatchIdentity(identity v1.Identity) bool
}

type IdentitySelectorImpl struct {
	IdentitySelector
}

func (i *IdentitySelectorImpl) MatchResource(list accessors.ElementListAccessor, a accessors.ResourceAccessor) bool {
	return i.MatchIdentity(a.GetMeta().GetIdentity(list))
}

func (i *IdentitySelectorImpl) MatchSource(list accessors.ElementListAccessor, a accessors.SourceAccessor) bool {
	return i.MatchIdentity(a.GetMeta().GetIdentity(list))
}

func (i *IdentitySelectorImpl) MatchReference(list accessors.ElementListAccessor, a accessors.ReferenceAccessor) bool {
	return i.MatchIdentity(a.GetMeta().GetIdentity(list))
}

type IdentityErrorSelectorImpl struct {
	ErrorSelectorBase
	IdentitySelectorImpl
}

func NewIdentityErrorSelectorImpl(s IdentitySelector, err error) *IdentityErrorSelectorImpl {
	return &IdentityErrorSelectorImpl{NewErrorSelectorBase(err), IdentitySelectorImpl{s}}
}

////////////////////////////////////////////////////////////////////////////////

type idattrs struct {
	v1.Identity
}

func (i *idattrs) MatchIdentity(identity v1.Identity) bool {
	for n, v := range i.Identity {
		if identity[n] != v {
			return false
		}
	}
	return true
}

func IdentityAttributesByKeyPairs(extra ...string) *IdentitySelectorImpl {
	return &IdentitySelectorImpl{&idattrs{v1.NewExtraIdentity(extra...)}}
}

func IdentityAttributes(identity v1.Identity) *IdentitySelectorImpl {
	return &IdentitySelectorImpl{&idattrs{identity}}
}

////////////////////////////////////////////////////////////////////////////////

type id struct {
	v1.Identity
}

func (i *id) MatchIdentity(identity v1.Identity) bool {
	if len(i.Identity) != len(identity) {
		return false
	}
	for n, v := range i.Identity {
		if identity[n] != v {
			return false
		}
	}
	return true
}

func IdentityByKeyPairs(name string, extra ...string) *IdentitySelectorImpl {
	return &IdentitySelectorImpl{&id{v1.NewIdentity(name, extra...)}}
}

func Identity(identity v1.Identity) *IdentitySelectorImpl {
	return &IdentitySelectorImpl{&id{identity}}
}

////////////////////////////////////////////////////////////////////////////////

type extraid struct {
	v1.Identity
}

func (i *extraid) MatchIdentity(identity v1.Identity) bool {
	identity = identity.ExtraIdentity()

	if len(i.Identity) != len(identity) {
		return false
	}
	for n, v := range i.Identity {
		if identity[n] != v {
			return false
		}
	}
	return true
}

func ExtraIdentityByKeyPairs(extra ...string) *IdentitySelectorImpl {
	return &IdentitySelectorImpl{&extraid{v1.NewExtraIdentity(extra...)}}
}

func ExtraIdentity(identity v1.Identity) *IdentitySelectorImpl {
	return &IdentitySelectorImpl{&extraid{identity}}
}

////////////////////////////////////////////////////////////////////////////////

type partialid struct {
	v1.Identity
}

func (i *partialid) MatchIdentity(identity v1.Identity) bool {
	for n, v := range i.Identity {
		if identity[n] != v {
			return false
		}
	}
	return true
}

func PartialIdentityByKeyPairs(attrs ...string) *IdentitySelectorImpl {
	return &IdentitySelectorImpl{&partialid{v1.NewExtraIdentity(attrs...)}}
}

func PartialIdentity(identity v1.Identity) *IdentitySelectorImpl {
	return &IdentitySelectorImpl{&partialid{identity}}
}

////////////////////////////////////////////////////////////////////////////////

type num int

func (i num) MatchIdentity(identity v1.Identity) bool {
	return len(identity) == int(i)
}

func NumberOfIdentityAttributes(n int) *IdentitySelectorImpl {
	return &IdentitySelectorImpl{num(n)}
}

////////////////////////////////////////////////////////////////////////////////

type idRegEx struct {
	attr string
	*regexp.Regexp
}

func (c *idRegEx) MatchIdentity(identity v1.Identity) bool {
	v, ok := identity[c.attr]
	if !ok {
		return false
	}
	return c.Regexp.MatchString(v)
}

func IdentityAttrRegex(name, ex string) *IdentitySelectorImpl {
	c, _ := regexp.Compile(ex)
	return &IdentitySelectorImpl{&idRegEx{name, c}}
}

////////////////////////////////////////////////////////////////////////////////

type Name string

func (n Name) MatchIdentity(identity v1.Identity) bool {
	return string(n) == identity[v1.SystemIdentityName]
}

func (n Name) MatchResource(list accessors.ElementListAccessor, r accessors.ResourceAccessor) bool {
	return string(n) == r.GetMeta().GetName()
}

func (n Name) MatchSource(list accessors.ElementListAccessor, r accessors.SourceAccessor) bool {
	return string(n) == r.GetMeta().GetName()
}

func (n Name) MatchReference(list accessors.ElementListAccessor, r accessors.ReferenceAccessor) bool {
	return string(n) == r.GetMeta().GetName()
}

////////////////////////////////////////////////////////////////////////////////

type Version string

func (v Version) MatchIdentity(identity v1.Identity) bool {
	return string(v) == identity[v1.SystemIdentityVersion]
}

func (v Version) MatchResource(list accessors.ElementListAccessor, r accessors.ResourceAccessor) bool {
	return string(v) == r.GetMeta().GetVersion()
}

func (v Version) MatchSource(list accessors.ElementListAccessor, r accessors.SourceAccessor) bool {
	return string(v) == r.GetMeta().GetVersion()
}

func (v Version) MatchReference(list accessors.ElementListAccessor, r accessors.ReferenceAccessor) bool {
	return string(v) == r.GetMeta().GetVersion()
}

////////////////////////////////////////////////////////////////////////////////

type semverConstraint struct {
	*semver.Constraints
}

func VersionConstraint(expr string) *semverConstraint {
	c, _ := semver.NewConstraint(expr)
	return &semverConstraint{c}
}

func (v *semverConstraint) check(vers string) bool {
	sv, _ := semver.NewVersion(vers)
	if sv == nil {
		return false
	}
	return v.Constraints.Check(sv)
}

func (v *semverConstraint) MatchIdentity(identity v1.Identity) bool {
	return v.check(identity[v1.SystemIdentityVersion])
}

func (v *semverConstraint) MatchResource(list accessors.ElementListAccessor, r accessors.ResourceAccessor) bool {
	return v.check(r.GetMeta().GetVersion())
}

func (v *semverConstraint) MatchSource(list accessors.ElementListAccessor, r accessors.SourceAccessor) bool {
	return v.check(r.GetMeta().GetVersion())
}

func (v *semverConstraint) MatchReference(list accessors.ElementListAccessor, r accessors.ReferenceAccessor) bool {
	return v.check(r.GetMeta().GetVersion())
}
