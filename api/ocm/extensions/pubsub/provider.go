package pubsub

import (
	"maps"
	"sync"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/general"

	"ocm.software/ocm/api/ocm/cpi"
)

// ProviderRegistry holds handlers able to extract
// a PubSub specification for an OCM repository of a dedicated kind.
type ProviderRegistry interface {
	Register(repoKind string, prov Provider)
	KnownProviders() map[string]Provider
	AddKnownProviders(registry ProviderRegistry)

	For(repo string) Provider
}

var DefaultRegistry = NewProviderRegistry()

func RegisterProvider(repokind string, prov Provider) {
	DefaultRegistry.Register(repokind, prov)
}

// A Provider is able to extract a pub sub configuration for
// an ocm repository (typically registered for a dedicated type of repository).
// It does not handle the pub sub system, but just the persistence of
// a pub sub specification configured for a dedicated type of repository.
type Provider interface {
	GetPubSubSpec(repo cpi.Repository) (PubSubSpec, error)
	SetPubSubSpec(repo cpi.Repository, spec PubSubSpec) error
}

type NopProvider struct{}

func (p NopProvider) GetPubSubSpec(repo cpi.Repository) (PubSubSpec, error) {
	return nil, nil
}

func (p NopProvider) SetPubSubSpec(repo cpi.Repository, spec PubSubSpec) error {
	return errors.ErrNotSupported("pub/sub configuration")
}

func NewProviderRegistry(base ...ProviderRegistry) ProviderRegistry {
	return &providers{
		base:      general.Optional(base...),
		providers: map[string]Provider{},
	}
}

type providers struct {
	lock sync.Mutex

	base      ProviderRegistry
	providers map[string]Provider
}

func (p *providers) Register(repoKind string, prov Provider) {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.providers[repoKind] = prov
}

func (p *providers) For(repo string) Provider {
	p.lock.Lock()
	defer p.lock.Unlock()
	prov := p.providers[repo]
	if prov != nil {
		return prov
	}
	if p.base != nil {
		return p.base.For(repo)
	}
	return nil
}

func (p *providers) KnownProviders() map[string]Provider {
	if p.base != nil {
		m := p.base.KnownProviders()
		for n, e := range p.providers {
			if m[n] == nil {
				m[n] = e
			}
		}
		return m
	}
	return maps.Clone(p.providers)
}

func (p *providers) AddKnownProviders(base ProviderRegistry) {
	for n, e := range base.KnownProviders() {
		p.providers[n] = e
	}
}
