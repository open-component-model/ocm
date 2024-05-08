package builder

import (
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/composition"
)

const T_OCM_COMPOSITION = "ocm composition repositoryt"

func (b *Builder) OCMCompositionRepository(name string, f ...func()) {
	r := composition.NewRepository(b, name)
	b.configure(&ocmRepository{Repository: r, kind: T_OCM_COMPOSITION}, f)
}
