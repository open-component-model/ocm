package resolvers

import (
	"github.com/mandelsoft/goutils/errors"
	"golang.org/x/exp/maps"

	"ocm.software/ocm/api/ocm/internal"
	common "ocm.software/ocm/api/utils/misc"
)

type (
	ContextProvider                  = internal.ContextProvider
	RepositorySpec                   = internal.RepositorySpec
	VersionLookup                    = internal.VersionLookup
	ComponentVersionAccess           = internal.ComponentVersionAccess
	ComponentVersionResolver         = internal.ComponentVersionResolver
	ComponentResolver                = internal.ComponentResolver
	Repository                       = internal.Repository
	ResolvedComponentVersionProvider = internal.ResolvedComponentVersionProvider
	ResolvedComponentProvider        = internal.ResolvedComponentProvider
	ResolverRule                     = internal.ResolverRule
)

const (
	KIND_COMPONENTVERSION = internal.KIND_COMPONENTVERSION
	KIND_COMPONENT        = internal.KIND_COMPONENT
	KIND_OCM_REFERENCE    = internal.KIND_OCM_REFERENCE
)

func NewResolverRule(prefix string, spec RepositorySpec, prio ...int) ResolverRule {
	return internal.NewResolverRule(prefix, spec, prio...)
}

// VersionResolverForComponent provides a VersionLookup for a component resolver.
// It resolves all versions provided by a component known to a ComponentResolver.
// The version set may be composed by versions of the component found in
// multiple repositories according to the result of the ComponentResolver.
func VersionResolverForComponent(name string, resolver ComponentResolver) (VersionLookup, error) {
	crs := resolver.LookupComponentProviders(name)
	if len(crs) == 0 {
		return nil, errors.ErrNotFound(KIND_COMPONENT, name)
	}

	versions := map[string]ResolvedComponentProvider{}
	for _, cr := range crs {
		c, err := cr.LookupComponent(name)
		if err != nil {
			return nil, err
		}
		vers, err := c.ListVersions()
		if err != nil {
			return nil, err
		}
		for _, v := range vers {
			if _, ok := versions[v]; !ok {
				versions[v] = cr
			}
		}
	}
	return &versionResolver{name, versions}, nil
}

type versionResolver struct {
	comp     string
	versions map[string]ResolvedComponentProvider
}

func (v *versionResolver) ListVersions() ([]string, error) {
	return maps.Keys(v.versions), nil
}

func (v *versionResolver) LookupVersion(version string) (ComponentVersionAccess, error) {
	p := v.versions[version]
	if p == nil {
		return nil, errors.ErrNotFound(KIND_COMPONENTVERSION, common.NewNameVersion(v.comp, version).String())
	}
	vp, err := p.LookupComponent(v.comp)
	if err != nil {
		return nil, err
	}
	return vp.LookupVersion(version)
}

func (v *versionResolver) HasVersion(vers string) (bool, error) {
	cv, err := v.LookupVersion(vers)
	if err != nil {
		return false, err
	}
	defer cv.Close()
	return cv != nil, nil
}
