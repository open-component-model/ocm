// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"github.com/open-component-model/ocm/pkg/contexts/config"
	cfgcpi "github.com/open-component-model/ocm/pkg/contexts/config/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/valuemergehandler/hpi"
	"github.com/open-component-model/ocm/pkg/runtime"
)

const (
	ConfigType   = "merge" + cfgcpi.OCM_CONFIG_TYPE_SUFFIX
	ConfigTypeV1 = ConfigType + runtime.VersionSeparator + "v1"
)

func init() {
	cfgcpi.RegisterConfigType(cfgcpi.NewConfigType[*Config](ConfigType, usage))
	cfgcpi.RegisterConfigType(cfgcpi.NewConfigType[*Config](ConfigTypeV1))
}

// Config describes a memory based config interface.
type Config struct {
	runtime.ObjectVersionedType `json:",inline"`
	Labels                      []LabelAssignment
	Assignments                 map[string]*hpi.Specification `json:"assignments,omitempty"`
}

type LabelAssignment struct {
	Name    string            `json:"name"`
	Version string            `json:"version,omitempty"`
	Merge   hpi.Specification `json:"merge,omitempty"`
}

// New creates a new memory ConfigSpec.
func New() *Config {
	return &Config{
		ObjectVersionedType: runtime.NewVersionedTypedObject(ConfigType),
		Assignments:         map[string]*hpi.Specification{},
	}
}

func (a *Config) GetType() string {
	return ConfigType
}

func (a *Config) Assign(name string, spec *hpi.Specification) {
	if a.Assignments == nil {
		a.Assignments = map[string]*hpi.Specification{}
	}
	if spec == nil {
		delete(a.Assignments, name)
	} else {
		a.Assignments[name] = spec
	}
}

func (a *Config) AssignLabel(name string, version string, spec *hpi.Specification) {
	if spec == nil {
		for i, s := range a.Labels {
			if s.Name == name && s.Version == version {
				a.Labels = append(a.Labels[:i], a.Labels[i+1:]...)
				return
			}
		}
	} else {
		a.Labels = append(a.Labels, LabelAssignment{
			Name:    name,
			Version: version,
			Merge:   *spec,
		})
	}
}

func (a *Config) ApplyTo(ctx config.Context, target interface{}) error {
	var reg hpi.Registry

	t, ok := target.(cpi.Context)
	if !ok {
		reg, ok = target.(hpi.Registry)
		if !ok {
			return config.ErrNoContext(ConfigType)
		}
	} else {
		reg = t.LabelMergeHandlers()
	}

	for n, s := range a.Assignments {
		reg.AssignHandler(n, s)
	}
	for _, s := range a.Labels {
		if s.Name == "" {
			continue
		}
		reg.AssignHandler(hpi.LabelHint(s.Name, s.Version), &s.Merge)
	}
	return nil
}

const usage = `
The config type <code>` + ConfigType + `</code> can be used to set some
assignments for the merging of (label) values. It applies to a value
merge handler registry, either directly or via an OCM context.

<pre>
    type: ` + ConfigType + `
    labels:
    - name: acme.org/audit/level
      merge:
        algorithm: acme.org/audit
        config: ...
    assignments:
       label:acme.org/audit/level@v1: 
          algorithm: acme.org/audit
          config: ...
          ...
</pre>
`
