package selectors

import (
	"github.com/open-component-model/ocm/pkg/contexts/ocm/selectors/accessors"
)

type ArtifactSelector interface {
	MatchArtifact(a accessors.ArtifactAccessor) bool
}

type ArtifactSelectorImpl struct {
	ArtifactSelector
}

func (i *ArtifactSelectorImpl) MatchResource(list accessors.ElementListAccessor, a accessors.ResourceAccessor) bool {
	return i.MatchArtifact(a)
}

func (i *ArtifactSelectorImpl) MatchSource(list accessors.ElementListAccessor, a accessors.SourceAccessor) bool {
	return i.MatchArtifact(a)
}

type ArtifactErrorSelectorImpl struct {
	ErrorSelectorBase
	ArtifactSelectorImpl
}

func NewArtifactErrorSelectorImpl(s ArtifactSelector, err error) *ArtifactErrorSelectorImpl {
	return &ArtifactErrorSelectorImpl{NewErrorSelectorBase(err), ArtifactSelectorImpl{s}}
}

////////////////////////////////////////////////////////////////////////////////

type artType string

func (n artType) MatchArtifact(a accessors.ArtifactAccessor) bool {
	return string(n) == a.GetType()
}

func ArtifactType(n string) *ArtifactSelectorImpl {
	return &ArtifactSelectorImpl{artType(n)}
}

////////////////////////////////////////////////////////////////////////////////

type accessKind string

func (n accessKind) MatchArtifact(a accessors.ArtifactAccessor) bool {
	return string(n) == a.GetAccess().GetKind()
}

func AccessKind(n string) *ArtifactSelectorImpl {
	return &ArtifactSelectorImpl{accessKind(n)}
}
