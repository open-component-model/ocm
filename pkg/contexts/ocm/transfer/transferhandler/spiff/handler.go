// Copyright 2022 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package spiff

import (
	"encoding/json"
	"strconv"

	"github.com/mandelsoft/spiff/dynaml"
	"github.com/mandelsoft/spiff/features"
	"github.com/mandelsoft/spiff/spiffing"
	"github.com/mandelsoft/spiff/yaml"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/transfer/transferhandler"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/transfer/transferhandler/standard"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/runtime"
)

type Handler struct {
	standard.Handler
	opts  *Options
	spiff spiffing.Spiff
}

func New(opts ...transferhandler.TransferOption) (transferhandler.TransferHandler, error) {
	options := &Options{}
	err := transferhandler.ApplyOptions(options, opts...)
	if err != nil {
		return nil, err
	}
	spiff := spiffing.New().WithFeatures(features.CONTROL, features.INTERPOLATION)
	if options.GetScriptFilesystem() != nil {
		spiff = spiff.WithFileSystem(options.fs)
	}
	return &Handler{
		Handler: *standard.NewDefaultHandler(&options.Options),
		opts:    options,
		spiff:   spiff,
	}, nil
}

func (h *Handler) TransferVersion(repo ocm.Repository, src ocm.ComponentVersionAccess, meta *compdesc.ElementMeta) (ocm.Repository, transferhandler.TransferHandler, error) {
	if h.opts.IsRecursive() {
		binding := h.getBinding(src, nil, meta)
		result, r, s, err := h.EvalRecursion("componentversion", binding, "process")
		if err != nil {
			return nil, nil, err
		}
		if result {
			if r != nil {
				repo, err = repo.GetContext().RepositoryForConfig(r, runtime.DefaultJSONEncoding)
				if err != nil {
					return nil, nil, err
				}
			}
			if s == nil {
				return repo, h, nil
			}
			opts := *h.opts
			opts.script = s
			return repo, &Handler{
				Handler: h.Handler,
				opts:    &opts,
			}, nil
		}
	}
	return nil, nil, nil
}

func (h *Handler) TransferResource(src ocm.ComponentVersionAccess, a ocm.AccessSpec, r ocm.ResourceAccess) (bool, error) {
	if !h.opts.IsResourcesByValue() {
		return false, nil
	}
	if h.opts.GetScript() == nil {
		return true, nil
	}
	binding := h.getBinding(src, a, &r.Meta().ElementMeta)
	return h.EvalBool("resource", binding, "process")
}

func (h *Handler) TransferSource(src ocm.ComponentVersionAccess, a ocm.AccessSpec, r ocm.SourceAccess) (bool, error) {
	if !h.opts.IsSourcesByValue() {
		return false, nil
	}
	if h.opts.GetScript() == nil {
		return true, nil
	}
	binding := h.getBinding(src, a, &r.Meta().ElementMeta)
	return h.EvalBool("source", binding, "process")
}

func (h *Handler) getBinding(src ocm.ComponentVersionAccess, a ocm.AccessSpec, m *compdesc.ElementMeta) map[string]interface{} {
	binding := map[string]interface{}{}
	binding["component"] = getCVAttrs(src)

	if a != nil {
		binding["access"] = getData(a)
	}
	binding["element"] = getData(m)
	return binding
}

func getData(in interface{}) interface{} {
	var v interface{}
	d, _ := json.Marshal(in)
	json.Unmarshal(d, &v)
	return v
}

func getCVAttrs(cv ocm.ComponentVersionAccess) map[string]interface{} {
	values := map[string]interface{}{}
	values["name"] = cv.GetName()
	values["version"] = cv.GetVersion()
	values["provider"] = string(cv.GetDescriptor().Provider)
	labels := map[string]interface{}{}
	for _, l := range cv.GetDescriptor().Labels {
		var m interface{}
		json.Unmarshal(l.Value, &m)
		labels[l.Name] = m
	}
	values["labels"] = labels
	return values
}

func (h *Handler) Eval(binding map[string]interface{}) (spiffing.Node, error) {
	spiff, err := h.spiff.WithValues(binding)
	if err != nil {
		return nil, err
	}
	node, err := spiff.Unmarshal("script", h.opts.GetScript())
	if err != nil {
		return nil, err
	}
	return spiff.Cascade(node, nil)
}

func (h *Handler) EvalBool(mode string, binding map[string]interface{}, key string) (bool, error) {
	binding = map[string]interface{}{
		"mode":   mode,
		"values": binding,
	}
	r, err := h.Eval(binding)
	if err != nil {
		return false, err
	}
	return h.evalBool(r, key)
}

func (h *Handler) EvalRecursion(mode string, binding map[string]interface{}, key string) (bool, []byte, []byte, error) {
	binding = map[string]interface{}{
		"mode":   mode,
		"values": binding,
	}
	r, err := h.Eval(binding)
	if err != nil {
		return false, nil, nil, err
	}

	m, ok := r.Value().(map[string]spiffing.Node)
	if !ok {
		return false, nil, nil, errors.ErrUnknown("field", key)
	}
	r = m[key]
	if r == nil {
		return false, nil, nil, errors.ErrUnknown("field", key)
	}

	b, err := h.evalBoolValue(r)
	if err == nil {
		// flat boolean without result structure
		return b, nil, nil, nil
	}
	m, ok = r.Value().(map[string]spiffing.Node)
	if !ok {
		return false, nil, nil, errors.ErrInvalid("result field type", dynaml.ExpressionType(r))
	}
	// now we expect a result structure
	// process: bool
	// repospec: map
	// script: template
	b, err = h.evalBool(r, key)
	if err != nil || !b {
		return false, nil, nil, err
	}
	var script []byte
	v := m["script"]
	if v != nil && v.Value() != nil {
		if t, ok := v.Value().(dynaml.TemplateValue); ok {
			if m, ok := t.Orig.Value().(map[string]spiffing.Node); ok {
				delete(m, "<<")
				delete(m, "<<<")
			} else {
				return false, nil, nil, errors.ErrInvalid("script template type", dynaml.ExpressionType(t.Orig))
			}
			script, err = yaml.Marshal(t.Orig)
			if err != nil {
				return false, nil, nil, err
			}
		} else {
			return false, nil, nil, errors.ErrInvalid("script type", dynaml.ExpressionType(v))
		}
	}

	var repospec []byte
	v = m["repospec"]
	if v != nil && v.Value() != nil {
		if _, ok := v.Value().(map[string]spiffing.Node); ok {
			spec, err := yaml.Normalize(v)
			if err == nil {
				repospec, err = json.Marshal(spec)
			}
			if err != nil {
				return false, nil, nil, errors.Wrapf(err, "invalid field repospec")
			}
		} else {
			return false, nil, nil, errors.ErrInvalid("repospec type", dynaml.ExpressionType(v))
		}
	}
	return true, repospec, script, nil
}

func (h *Handler) evalBool(r spiffing.Node, key string) (bool, error) {
	v := r.Value().(map[string]spiffing.Node)[key]
	if v == nil {
		return false, errors.ErrUnknown("field", key)
	}
	return h.evalBoolValue(v)
}

func (h *Handler) evalBoolValue(v spiffing.Node) (bool, error) {
	switch b := v.Value().(type) {
	case bool:
		return b, nil
	case int64:
		return b != 0, nil
	case float64:
		return b != 0, nil
	case string:
		r, err := strconv.ParseBool(b)
		if err != nil {
			return len(b) > 0, nil
		}
		return r, nil
	default:
		return false, errors.ErrInvalid("boolean result", dynaml.ExpressionType(v))
	}
}
