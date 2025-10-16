package plugin

import (
	"bytes"
	"encoding/json"

	"github.com/mandelsoft/goutils/errors"
	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/credentials/identity/hostpath"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/cpi/accspeccpi"
	"ocm.software/ocm/api/ocm/plugin"
	"ocm.software/ocm/api/ocm/plugin/descriptor"
	"ocm.software/ocm/api/ocm/plugin/ppi"
	"ocm.software/ocm/api/utils/errkind"
)

type plug = plugin.Plugin

// PluginHandler is a shared object between the AccessMethod implementation and the AccessSpec implementation. The
// object knows the actual plugin and can therefore forward the method calls to corresponding cli commands.
type PluginHandler struct {
	plug

	// cached info
	info *ppi.AccessSpecInfo
	err  error
	orig []byte
}

func NewPluginHandler(p plugin.Plugin) *PluginHandler {
	return &PluginHandler{plug: p}
}

func (p *PluginHandler) Info(spec *AccessSpec) (*ppi.AccessSpecInfo, error) {
	if p.info != nil || p.err != nil {
		raw, err := spec.UnstructuredVersionedTypedObject.GetRaw()
		if err != nil {
			return nil, errors.Wrapf(err, "cannot marshal access specification")
		}
		if bytes.Equal(raw, p.orig) {
			return p.info, p.err
		}
	}
	p.info, p.err = p.Validate(spec)
	return p.info, p.err
}

func (p *PluginHandler) AccessMethod(spec *AccessSpec, cv cpi.ComponentVersionAccess) (cpi.AccessMethod, error) {
	mspec := p.GetAccessMethodDescriptor(spec.GetKind(), spec.GetVersion())
	if mspec == nil {
		return nil, errors.ErrNotFound(errkind.KIND_ACCESSMETHOD, spec.GetType(), descriptor.KIND_PLUGIN, p.Name())
	}

	creddata, err := p.getCredentialData(spec, cv)
	if err != nil {
		return nil, err
	}

	info, err := p.Info(spec)
	if err != nil {
		return nil, err
	}
	return accspeccpi.AccessMethodForImplementation(newMethod(p, spec, cv.GetContext(), info, creddata), nil)
}

func (p *PluginHandler) getCredentialData(spec *AccessSpec, cv cpi.ComponentVersionAccess) (json.RawMessage, error) {
	info, err := p.Info(spec)
	if err != nil {
		return nil, err
	}

	var creds credentials.Credentials
	if len(info.ConsumerId) > 0 {
		creds, err = credentials.CredentialsForConsumer(cv.GetContext(), info.ConsumerId, hostpath.IdentityMatcher(info.ConsumerId.Type()))
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

func (p *PluginHandler) Describe(spec *AccessSpec, ctx cpi.Context) string {
	mspec := p.GetAccessMethodDescriptor(spec.GetKind(), spec.GetVersion())
	if mspec == nil {
		return "unknown type " + spec.GetType()
	}
	info, err := p.Info(spec)
	if err != nil {
		return err.Error()
	}
	return info.Short
}

func (p *PluginHandler) GetMimeType(spec *AccessSpec) string {
	mspec := p.GetAccessMethodDescriptor(spec.GetKind(), spec.GetVersion())
	if mspec == nil {
		return "unknown type " + spec.GetType()
	}
	info, err := p.Info(spec)
	if err != nil {
		return ""
	}
	return info.Short
}

func (p *PluginHandler) GetReferenceHint(spec *AccessSpec, cv cpi.ComponentVersionAccess) string {
	mspec := p.GetAccessMethodDescriptor(spec.GetKind(), spec.GetVersion())
	if mspec == nil {
		return "unknown type " + spec.GetType()
	}
	info, err := p.Info(spec)
	if err != nil {
		return ""
	}
	return info.Hint
}

func (p *PluginHandler) Validate(spec *AccessSpec) (*ppi.AccessSpecInfo, error) {
	data, err := spec.GetRaw()
	if err != nil {
		return nil, err
	}
	return p.plug.ValidateAccessMethod(data)
}
