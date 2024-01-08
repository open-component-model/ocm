// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package certattr

import (
	"sync"

	"github.com/mandelsoft/vfs/pkg/vfs"
	cfgcpi "github.com/open-component-model/ocm/pkg/contexts/config/cpi"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/runtime"
	"github.com/open-component-model/ocm/pkg/signing/signutils"
	"github.com/open-component-model/ocm/pkg/utils"
)

const (
	ConfigType   = "rootcerts" + cfgcpi.OCM_CONFIG_TYPE_SUFFIX
	ConfigTypeV1 = ConfigType + runtime.VersionSeparator + "v1"
)

func init() {
	cfgcpi.RegisterConfigType(cfgcpi.NewConfigType[*Config](ConfigType, usage))
	cfgcpi.RegisterConfigType(cfgcpi.NewConfigType[*Config](ConfigTypeV1, usage))
}

// Config describes a memory based repository interface.
type Config struct {
	runtime.ObjectVersionedType `json:",inline"`
	RootCertificates            []cfgcpi.ContentSpec `json:"rootCertificates,omitempty"`
}

// New creates a new memory ConfigSpec.
func New() *Config {
	return &Config{
		ObjectVersionedType: runtime.NewVersionedTypedObject(ConfigType),
	}
}

func (a *Config) GetType() string {
	return ConfigType
}

func (a *Config) AddRootCertifacteFile(name string, fss ...vfs.FileSystem) {
	a.RootCertificates = append(a.RootCertificates, cfgcpi.ContentSpec{Path: name, FileSystem: utils.Optional(fss...)})
}

func (a *Config) AddRootCertifacteData(data []byte) {
	a.RootCertificates = append(a.RootCertificates, cfgcpi.ContentSpec{Data: data})
}

func (a *Config) AddRootCertifacte(chain signutils.GenericCertificateChain) error {
	certs, err := signutils.GetCertificateChain(chain, false)
	if err != nil {
		return err
	}
	a.RootCertificates = append(a.RootCertificates, cfgcpi.ContentSpec{Data: signutils.CertificateChainToPem(certs), Parsed: certs})
	return nil
}

func (a *Config) ApplyTo(ctx cfgcpi.Context, target interface{}) error {
	switch t := target.(type) {
	case Context:
		errors.Wrapf(a.ApplyToAttribute(Get(t)), "applying config to certattr failed")
	}
	return cfgcpi.ErrNoContext(ConfigType)
}

func (a *Config) ApplyToAttribute(attr *Attribute) error {
	for i, k := range a.RootCertificates {
		key, err := k.Get()
		if err != nil {
			return errors.Wrapf(err, "cannot get root certificate %d", i)
		}
		err = attr.RegisterRootCertificates(key)
		if err != nil {
			return errors.Wrapf(err, "invalid certificate %d", i)
		}
	}
	return nil
}

const usage = `
The config type <code>` + ConfigType + `</code> can be used to define
general root certificates. A certificate value might be given by one of the fields:
- <code>path</code>: path of file with key data
- <code>data</code>: base64 encoded binary data
- <code>stringdata</code>: data a string parsed by key handler

<pre>
    rootCertificates:
      - path: &lt;file path>
</pre>

`

type Appliers struct {
	lock     sync.Mutex
	appliers []cfgcpi.ConfigApplier
}

func (r *Appliers) Register(a ...cfgcpi.ConfigApplier) {
	r.lock.Lock()
	defer r.lock.Unlock()

	r.appliers = append(r.appliers, a...)
}

var DefaultAppliers = &Appliers{}

func RegisterApplier(a ...cfgcpi.ConfigApplier) {
	DefaultAppliers.Register(a...)
}
