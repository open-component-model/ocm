package srcsel

import (
	"github.com/open-component-model/ocm/pkg/contexts/ocm/selectors"
)

// Artifact selectors

func ArtifactType(n string) Selector {
	return selectors.ArtifactType(n)
}

func AccessKind(n string) Selector {
	return selectors.AccessKind(n)
}
