// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"fmt"

	"github.com/open-component-model/ocm/pkg/contexts/clictx/internal"
	"github.com/open-component-model/ocm/pkg/contexts/config"
	"github.com/open-component-model/ocm/pkg/contexts/config/cpi"
	ocicpi "github.com/open-component-model/ocm/pkg/contexts/oci/cpi"
	ocmcpi "github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/runtime"
)

const (
	OCMCmdConfigType   = "ocm.cmd" + cpi.OCM_CONFIG_TYPE_SUFFIX
	OCMCmdConfigTypeV1 = OCMCmdConfigType + runtime.VersionSeparator + "v1"
)

func init() {
	cpi.RegisterConfigType(OCMCmdConfigType, cpi.NewConfigType(OCMCmdConfigType, &Config{}, usage))
	cpi.RegisterConfigType(OCMCmdConfigTypeV1, cpi.NewConfigType(OCMCmdConfigTypeV1, &Config{}, usage))
}

// Config describes a memory based repository interface.
type Config struct {
	runtime.ObjectVersionedType `json:",inline"`
	OCMRepositories             map[string]*ocmcpi.GenericRepositorySpec `json:"ocmRepositories,omitempty"`
	OCIRepositories             map[string]*ocicpi.GenericRepositorySpec `json:"ociRepositories,omitempty"`
}

// New creates a new memory ConfigSpec.
func New() *Config {
	return &Config{
		ObjectVersionedType: runtime.NewVersionedTypedObject(OCMCmdConfigType),
	}
}

func (a *Config) GetType() string {
	return OCMCmdConfigType
}

func (a *Config) AddOCIRepository(name string, spec ocicpi.RepositorySpec) error {
	g, err := ocicpi.ToGenericRepositorySpec(spec)
	if err != nil {
		return fmt.Errorf("unable to convert oci repository spec to generic spec: %w", err)
	}

	if a.OCIRepositories == nil {
		a.OCIRepositories = map[string]*ocicpi.GenericRepositorySpec{}
	}

	a.OCIRepositories[name] = g

	return nil
}

func (a *Config) AddOCMRepository(name string, spec ocmcpi.RepositorySpec) error {
	g, err := ocmcpi.ToGenericRepositorySpec(spec)
	if err != nil {
		return fmt.Errorf("unable to convert ocm repository spec to generic spec: %w", err)
	}

	if a.OCMRepositories == nil {
		a.OCMRepositories = map[string]*ocmcpi.GenericRepositorySpec{}
	}

	a.OCMRepositories[name] = g

	return nil
}

func (a *Config) ApplyTo(ctx config.Context, target interface{}) error {
	t, ok := target.(internal.Context)
	if !ok {
		return config.ErrNoContext(OCMCmdConfigType)
	}
	for n, s := range a.OCIRepositories {
		t.OCI().Context().SetAlias(n, s)
	}
	for n, s := range a.OCMRepositories {
		t.OCM().Context().SetAlias(n, s)
	}
	return nil
}

const usage = `
The config type <code>` + OCMCmdConfigType + `</code> can be used to 
configure predefined aliases for dedicated OCM repositories and 
OCI registries.

<pre>
   type: ` + OCMCmdConfigType + `
   ocmRepositories:
   &lt;name>: &lt;specification of OCM repository>
   ...
   ociRepositories:
   &lt;name>: &lt;specification of OCI registry>
   ...
</pre>
`
