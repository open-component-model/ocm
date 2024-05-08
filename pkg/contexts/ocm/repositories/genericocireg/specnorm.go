package genericocireg

import (
	"slices"
	"sync"
)

type SpecificationNormalizer func(s *RepositorySpec)

type Normalizers struct {
	lock     sync.Mutex
	handlers map[string][]SpecificationNormalizer
}

func (n *Normalizers) Register(typ string, f SpecificationNormalizer) {
	n.lock.Lock()
	defer n.lock.Unlock()

	n.handlers[typ] = append(n.handlers[typ], f)
}

func (n *Normalizers) Get(typ string) []SpecificationNormalizer {
	n.lock.Lock()
	defer n.lock.Unlock()
	return slices.Clone(n.handlers[typ])
}

func (n *Normalizers) Normalize(s *RepositorySpec) *RepositorySpec {
	n.lock.Lock()
	defer n.lock.Unlock()

	found := false
	for _, f := range n.handlers[s.GetType()] {
		found = true
		f(s)
	}
	if !found && s.GetType() != s.GetKind() {
		for _, f := range n.handlers[s.GetKind()] {
			f(s)
		}
	}
	return s
}

var normalizers = &Normalizers{handlers: map[string][]SpecificationNormalizer{}}

// RegisterSpecificationNormalizer can be used to register OCI repository type
// specific handlers used to normalize an OCI type based OCM repository spec.
func RegisterSpecificationNormalizer(typ string, f SpecificationNormalizer) {
	normalizers.Register(typ, f)
}
