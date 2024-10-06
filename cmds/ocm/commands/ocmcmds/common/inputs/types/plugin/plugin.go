package plugin

import (
	"bytes"
	"encoding/json"
	"sync"

	"github.com/mandelsoft/goutils/errors"

	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/credentials/identity/hostpath"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/plugin"
	"ocm.software/ocm/api/ocm/plugin/descriptor"
)

type plug = plugin.Plugin

// PluginHandler is a shared object between the GetAccess implementation and the Spec implementation. The
// object knows the actual plugin and can therefore forward the method calls to corresponding cli commands.
type PluginHandler struct {
	lock sync.Mutex
	plug

	// cached info
	access *access
	err    error
	orig   []byte
}

func NewPluginHandler(p plugin.Plugin) *PluginHandler {
	return &PluginHandler{plug: p}
}

func (p *PluginHandler) GetAccess(spec *Spec, ctx cpi.Context) (*access, error) {
	raw, err := spec.GetRaw()
	if err != nil {
		return nil, errors.Wrapf(err, "cannot marshal input specification")
	}
	p.lock.Lock()
	defer p.lock.Unlock()

	if p.access != nil || p.err != nil {
		if bytes.Equal(raw, p.orig) {
			return p.access, p.err
		}
	}
	mspec := p.GetInputTypeDescriptor(spec.GetKind(), spec.GetVersion())
	if mspec == nil {
		return nil, errors.ErrNotFound(descriptor.KIND_INPUTTYPE, spec.GetType(), descriptor.KIND_PLUGIN, p.Name())
	}

	info, err := p.plug.ValidateInputSpec(raw)
	p.err = err
	if err != nil {
		return nil, err
	}
	p.access = newAccess(p, spec, ctx, info)

	creddata, err := p.getCredentialData(info, ctx)
	if err != nil {
		return nil, err
	}
	p.access.creds = creddata

	return p.access, nil
}

func (p *PluginHandler) getCredentialData(info *plugin.InputSpecInfo, ctx cpi.Context) (json.RawMessage, error) {
	var (
		err   error
		creds credentials.Credentials
	)

	if len(info.ConsumerId) > 0 {
		creds, err = credentials.CredentialsForConsumer(ctx, info.ConsumerId, hostpath.IdentityMatcher(info.ConsumerId.Type()))
		if err != nil {
			return nil, err
		}
	}

	var creddata json.RawMessage
	if creds != nil {
		creddata, err = json.Marshal(creds)
		if err != nil {
			return nil, err
		}
	}
	return creddata, nil
}

func (p *PluginHandler) Describe(spec *Spec, ctx cpi.Context) string {
	acc, err := p.GetAccess(spec, ctx)
	if err != nil {
		return err.Error()
	}
	return acc.info.Short
}
