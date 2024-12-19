package featuregates

import (
	"sync"

	"github.com/mandelsoft/goutils/general"
	"github.com/mandelsoft/goutils/maputils"

	"ocm.software/ocm/api/datacontext"
	"ocm.software/ocm/api/datacontext/attrs/featuregatesattr"
	"ocm.software/ocm/api/utils"
	common "ocm.software/ocm/api/utils/misc"
)

type FeatureGate struct {
	Name        string `json:"name"`
	Short       string `json:"short"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`
}

func (f *FeatureGate) GetSettings(ctx datacontext.Context) *featuregatesattr.FeatureGate {
	return featuregatesattr.Get(ctx).GetFeature(f.Name, f.Enabled)
}

func (f *FeatureGate) IsEnabled(ctx datacontext.Context) bool {
	return featuregatesattr.Get(ctx).IsEnabled(f.Name, f.Enabled)
}

type Registry interface {
	Register(gate *FeatureGate)
	GetNames() []string
	Get(name string) *FeatureGate
}

type registry struct {
	lock  sync.Mutex
	gates map[string]*FeatureGate
}

var _ Registry = (*registry)(nil)

func (r *registry) Register(g *FeatureGate) {
	r.lock.Lock()
	defer r.lock.Unlock()

	r.gates[g.Name] = g
}

func (r *registry) GetNames() []string {
	r.lock.Lock()
	defer r.lock.Unlock()

	return maputils.OrderedKeys(r.gates)
}

func (r *registry) Get(name string) *FeatureGate {
	r.lock.Lock()
	defer r.lock.Unlock()

	return r.gates[name]
}

var defaultRegistry = &registry{
	gates: map[string]*FeatureGate{},
}

func DefaultRegistry() Registry {
	return defaultRegistry
}

func Register(fg *FeatureGate) {
	defaultRegistry.Register(fg)
}

func Usage(reg Registry) string {
	p, buf := common.NewBufferedPrinter()
	for _, n := range reg.GetNames() {
		a := reg.Get(n)
		p.Printf("- Name: %s\n", n)
		p.Printf("  Default: %s\n", general.Conditional(a.Enabled, "enabled", "disabled"))
		if a.Description != "" {
			p.Printf("%s\n", utils.IndentLines(a.Description, "    "))
		}
	}
	return buf.String()
}
