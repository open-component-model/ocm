package config

import (
	"fmt"

	"ocm.software/ocm/api/cli/internal"
	"ocm.software/ocm/api/config"
	"ocm.software/ocm/api/config/cpi"
	ocicpi "ocm.software/ocm/api/oci/cpi"
	ocmcpi "ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/utils/runtime"
)

const (
	OCMCmdConfigType   = "ocm.cmd" + cpi.OCM_CONFIG_TYPE_SUFFIX
	OCMCmdConfigTypeV1 = OCMCmdConfigType + runtime.VersionSeparator + "v1"
)

func init() {
	cpi.RegisterConfigType(cpi.NewConfigType[*Config](OCMCmdConfigType, usage))
	cpi.RegisterConfigType(cpi.NewConfigType[*Config](OCMCmdConfigTypeV1, usage))
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
