package compdesc

import (
	"sort"
	"sync"

	"github.com/mandelsoft/goutils/errors"
)

// NormalisationAlgorithm types and versions the algorithm used for digest generation.
type NormalisationAlgorithm = string

const (
	// Deprecated: use JsonNormalisationV3 instead
	JsonNormalisationV1 NormalisationAlgorithm = "jsonNormalisation/v1"
	// Deprecated: use JsonNormalisationV3 instead
	JsonNormalisationV2 NormalisationAlgorithm = "jsonNormalisation/v2"
	JsonNormalisationV3 NormalisationAlgorithm = "jsonNormalisation/v3"
)

type Normalization interface {
	Normalize(cd *ComponentDescriptor) ([]byte, error)
}

type NormalizationAlgorithms struct {
	sync.RWMutex
	algos map[string]Normalization
}

func (n *NormalizationAlgorithms) Register(name string, norm Normalization) {
	n.Lock()
	defer n.Unlock()
	n.algos[name] = norm
}

func (n *NormalizationAlgorithms) Get(algo string) Normalization {
	n.RLock()
	defer n.RUnlock()
	return n.algos[algo]
}

func (n *NormalizationAlgorithms) Names() []string {
	n.RLock()
	defer n.RUnlock()
	names := []string{}
	for n := range n.algos {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

func (n *NormalizationAlgorithms) Normalize(cd *ComponentDescriptor, algo string) ([]byte, error) {
	n.RLock()
	defer n.RUnlock()

	norm := n.algos[algo]
	if norm == nil {
		return nil, errors.ErrUnknown("normalization algorithm", algo)
	}
	return norm.Normalize(cd)
}

var Normalizations = NormalizationAlgorithms{algos: map[string]Normalization{}}

func Normalize(cd *ComponentDescriptor, normAlgo string) ([]byte, error) {
	DefaultComponent(cd)
	return Normalizations.Normalize(cd, normAlgo)
}
