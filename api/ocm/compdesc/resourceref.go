package compdesc

import (
	"fmt"

	"github.com/mandelsoft/goutils/errors"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	common "ocm.software/ocm/api/utils/misc"
)

func ResolveReferencePath(cv *ComponentDescriptor, path []metav1.Identity, resolver ComponentVersionResolver) (*ComponentDescriptor, error) {
	if cv == nil {
		return nil, fmt.Errorf("no component version specified")
	}

	eff := cv
	for _, cr := range path {
		cref, err := eff.GetReferenceByIdentity(cr)
		if err != nil {
			return nil, errors.Wrapf(err, "%s", common.VersionedElementKey(cv))
		}

		compoundResolver := NewCompoundResolver(NewComponentVersionSet(cv), resolver)
		eff, err = compoundResolver.LookupComponentVersion(cref.GetComponentName(), cref.GetVersion())
		if err != nil {
			return nil, errors.Wrapf(err, "cannot resolve component version for reference %s", cr.String())
		}
		if eff == nil {
			return nil, errors.ErrNotFound(KIND_COMPONENTVERSION, cref.String())
		}
	}
	return eff, nil
}

func MatchResourceReference(cv *ComponentDescriptor, typ string, ref metav1.ResourceReference, resolver ComponentVersionResolver) (*Resource, *ComponentDescriptor, error) {
	eff, err := ResolveReferencePath(cv, ref.ReferencePath, resolver)
	if err != nil {
		return nil, nil, err
	}

	if len(eff.Resources) == 0 && len(ref.Resource) == 0 {
		return nil, nil, errors.ErrNotFound(KIND_RESOURCE)
	}
outer:
	for i, r := range eff.Resources {
		if r.Type != typ && typ != "" {
			continue
		}
		for k, v := range ref.Resource {
			switch k {
			case metav1.SystemIdentityName:
				if v != r.Name {
					continue outer
				}
			case metav1.SystemIdentityVersion:
				if v != r.Version {
					continue outer
				}
			default:
				if r.ExtraIdentity == nil || r.ExtraIdentity[k] != v {
					continue outer
				}
			}
		}
		return &eff.Resources[i], eff, nil
	}
	return nil, nil, errors.ErrNotFound(KIND_RESOURCE, ref.Resource.String())
}

func ResolveResourceReference(cd *ComponentDescriptor, ref metav1.ResourceReference, resolver ComponentVersionResolver) (*Resource, *ComponentDescriptor, error) {
	if len(ref.Resource) == 0 || len(ref.Resource["name"]) == 0 {
		return nil, nil, errors.Newf("at least resource name must be specified for resource reference")
	}

	eff, err := ResolveReferencePath(cd, ref.ReferencePath, resolver)
	if err != nil {
		return nil, nil, err
	}
	r, err := eff.GetResourceByIdentity(ref.Resource)
	if err != nil {
		return nil, nil, err
	}
	return &r, eff, nil
}
