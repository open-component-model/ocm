package v3alpha1

import (
	"ocm.software/ocm/api/utils/runtime"
)

// Default applies defaults to a component.
func (cd *ComponentDescriptor) Default() error {
	if cd.RepositoryContexts == nil {
		cd.RepositoryContexts = make([]*runtime.UnstructuredTypedObject, 0)
	}
	if cd.Spec.Sources == nil {
		cd.Spec.Sources = make([]Source, 0)
	}
	if cd.Spec.References == nil {
		cd.Spec.References = make([]Reference, 0)
	}
	if cd.Spec.Resources == nil {
		cd.Spec.Resources = make([]Resource, 0)
	}

	return nil
}
