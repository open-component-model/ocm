package v2

import (
	"ocm.software/ocm/api/utils/runtime"
)

// Default applies defaults to a component.
func (cd *ComponentDescriptor) Default() error {
	if cd.RepositoryContexts == nil {
		cd.RepositoryContexts = make([]*runtime.UnstructuredTypedObject, 0)
	}
	if cd.Sources == nil {
		cd.Sources = make([]Source, 0)
	}
	if cd.ComponentReferences == nil {
		cd.ComponentReferences = make([]ComponentReference, 0)
	}
	if cd.Resources == nil {
		cd.Resources = make([]Resource, 0)
	}

	return nil
}
