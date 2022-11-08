// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package plugin

import (
	"bytes"
	"encoding/json"

	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	"github.com/open-component-model/ocm/pkg/contexts/credentials/identity/hostpath"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/ppi"
	"github.com/open-component-model/ocm/pkg/errors"
)

type plug = plugin.Plugin

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
		return nil, errors.ErrNotFound(errors.KIND_ACCESSMETHOD, spec.GetType(), ppi.KIND_PLUGIN, p.Name())
	}
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
			return nil, errors.Wrapf(err, "cannot marshal access spec")
		}
	}
	return newMethod(p, spec, cv.GetContext(), info, creddata), nil
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
