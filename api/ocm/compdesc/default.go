package compdesc

import (
	v1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/utils/runtime"
)

// DefaultComponent applies defaults to a component.
func DefaultComponent(component *ComponentDescriptor) *ComponentDescriptor {
	if component.RepositoryContexts == nil {
		component.RepositoryContexts = make([]*runtime.UnstructuredTypedObject, 0)
	}
	if component.Sources == nil {
		component.Sources = make([]Source, 0)
	}
	if component.References == nil {
		component.References = make([]Reference, 0)
	}
	if component.Resources == nil {
		component.Resources = make([]Resource, 0)
	}

	if component.Metadata.ConfiguredVersion == "" {
		component.Metadata.ConfiguredVersion = DefaultSchemeVersion
	}
	// DefaultResources(component)
	return component
}

func DefaultElements(component *ComponentDescriptor) {
	DefaultResources(component)
	DefaultSources(component)
	DefaultReferences(component)
}

// DefaultResources defaults a list of resources.
// The version of the component is defaulted for local resources that do not contain a version.
// adds the version as identity if the resource identity would clash otherwise.
// The version is added to an extraIdentity, if it is not unique without it.
func DefaultResources(component *ComponentDescriptor) {
	for i, res := range component.Resources {
		if res.Relation == v1.LocalRelation && len(res.Version) == 0 {
			component.Resources[i].Version = component.GetVersion()
		}

		id := res.GetIdentity(component.Resources)
		if v, ok := id[SystemIdentityVersion]; ok {
			if res.ExtraIdentity == nil {
				component.Resources[i].ExtraIdentity = v1.Identity{
					SystemIdentityVersion: v,
				}
			} else {
				if _, ok := res.ExtraIdentity[SystemIdentityVersion]; !ok {
					res.ExtraIdentity[SystemIdentityVersion] = v
				}
			}
		}
	}
}

// DefaultSources defaults a list of sources.
// The version is added to an extraIdentity, if it is not unique without it.
func DefaultSources(component *ComponentDescriptor) {
	for i, res := range component.Sources {
		id := res.GetIdentity(component.Resources)
		if v, ok := id[SystemIdentityVersion]; ok {
			if res.ExtraIdentity == nil {
				component.Sources[i].ExtraIdentity = v1.Identity{
					SystemIdentityVersion: v,
				}
			} else {
				if _, ok := res.ExtraIdentity[SystemIdentityVersion]; !ok {
					res.ExtraIdentity[SystemIdentityVersion] = v
				}
			}
		}
	}
}

// DefaultReferences defaults a list of references.
// The version is added to an extraIdentity, if it is not unique without it.
func DefaultReferences(component *ComponentDescriptor) {
	for i, res := range component.References {
		id := res.GetIdentity(component.Resources)
		if v, ok := id[SystemIdentityVersion]; ok {
			if res.ExtraIdentity == nil {
				component.References[i].ExtraIdentity = v1.Identity{
					SystemIdentityVersion: v,
				}
			} else {
				if _, ok := res.ExtraIdentity[SystemIdentityVersion]; !ok {
					res.ExtraIdentity[SystemIdentityVersion] = v
				}
			}
		}
	}
}
