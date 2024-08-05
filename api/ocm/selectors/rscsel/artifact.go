package rscsel

import (
	"ocm.software/ocm/api/ocm/selectors"
)

// Artifact selectors

func ArtifactType(n string) Selector {
	return selectors.ArtifactType(n)
}

func AccessKind(n string) Selector {
	return selectors.AccessKind(n)
}
