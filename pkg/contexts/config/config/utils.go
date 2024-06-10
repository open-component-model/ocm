package config

import (
	"github.com/open-component-model/ocm/pkg/contexts/config/cpi"
)

type Aggregator struct {
	cfg       cpi.Config
	aggr      *Config
	optimized bool
}

func NewAggregator(optimized bool, cfgs ...cpi.Config) (*Aggregator, error) {
	a := &Aggregator{optimized: optimized}
	for _, c := range cfgs {
		err := a.AddConfig(c)
		if err != nil {
			return nil, err
		}
	}
	return a, nil
}

func (a *Aggregator) Get() cpi.Config {
	return a.cfg
}

func (a *Aggregator) AddConfig(cfg cpi.Config) error {
	if a.cfg == nil {
		a.cfg = cfg
		if aggr, ok := cfg.(*Config); ok && a.optimized {
			a.aggr = aggr
		}
	} else {
		if a.aggr == nil {
			a.aggr = New()
			err := a.aggr.AddConfig(a.cfg)
			if err != nil {
				return err
			}
			cfg = a.aggr
		}
		err := a.aggr.AddConfig(cfg)
		if err != nil {
			return err
		}
	}
	return nil
}
