package spiff

import (
	"encoding/json"
	"strconv"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/spiff/dynaml"
	"github.com/mandelsoft/spiff/features"
	"github.com/mandelsoft/spiff/spiffing"
	"github.com/mandelsoft/spiff/yaml"
	"github.com/sirupsen/logrus"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/compdesc"
	"ocm.software/ocm/api/ocm/tools/transfer/transferhandler"
	"ocm.software/ocm/api/ocm/tools/transfer/transferhandler/standard"
	"ocm.software/ocm/api/utils/runtime"
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

// TODO: handle update and overwrite per script

func (h *Handler) UpdateVersion(src ocm.ComponentVersionAccess, tgt ocm.ComponentVersionAccess) (bool, error) {
	if ok, err := h.Handler.UpdateVersion(src, tgt); !ok || err != nil {
		return ok, nil
	}
	if h.opts.GetScript() == nil {
		return false, nil
	}
	binding := h.getBinding(src, nil, nil, nil, nil)
	return h.EvalBool("update", binding, "process")
}

func (h *Handler) EnforceTransport(src ocm.ComponentVersionAccess, tgt ocm.ComponentVersionAccess) (bool, error) {
	if ok, err := h.Handler.EnforceTransport(src, tgt); ok || err != nil {
		return ok, nil
	}
	if h.opts.GetScript() == nil {
		return false, nil
	}
	binding := h.getBinding(src, nil, nil, nil, nil)
	return h.EvalBool("enforceTransport", binding, "process")
}

func (h *Handler) OverwriteVersion(src ocm.ComponentVersionAccess, tgt ocm.ComponentVersionAccess) (bool, error) {
	if ok, err := h.Handler.OverwriteVersion(src, tgt); ok || err != nil {
		return ok, nil
	}
	if h.opts.GetScript() == nil {
		return false, nil
	}
	binding := h.getBinding(src, nil, nil, nil, nil)
	return h.EvalBool("overwrite", binding, "process")
}

func (h *Handler) TransferVersion(repo ocm.Repository, src ocm.ComponentVersionAccess, meta *compdesc.Reference, tgt ocm.Repository) (ocm.ComponentVersionAccess, transferhandler.TransferHandler, error) {
	if src == nil || h.opts.IsRecursive() {
		if h.opts.GetScript() == nil {
			return h.Handler.TransferVersion(repo, src, meta, tgt)
		}
		binding := h.getBinding(src, nil, &meta.ElementMeta, nil, tgt)
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
				return h.Handler.TransferVersion(repo, src, meta, tgt)
			}
			opts := *h.opts
			opts.script = s
			cv, _, err := h.Handler.TransferVersion(repo, src, meta, tgt)
			return cv, &Handler{
				Handler: h.Handler,
				opts:    &opts,
			}, err
		}
	}
	return nil, nil, nil
}

func (h *Handler) TransferResource(src ocm.ComponentVersionAccess, a ocm.AccessSpec, r ocm.ResourceAccess) (bool, error) {
	if !h.opts.IsResourcesByValue() {
		return false, nil
	}
	if h.opts.IsAccessTypeOmitted(a.GetType()) {
		return false, nil
	}
	if h.opts.GetScript() == nil {
		return true, nil
	}
	binding := h.getBinding(src, a, &r.Meta().ElementMeta, &r.Meta().Type, nil)
	return h.EvalBool("resource", binding, "process")
}

func (h *Handler) TransferSource(src ocm.ComponentVersionAccess, a ocm.AccessSpec, r ocm.SourceAccess) (bool, error) {
	if !h.opts.IsSourcesByValue() {
		return false, nil
	}
	if h.opts.IsAccessTypeOmitted(a.GetType()) {
		return false, nil
	}
	if h.opts.GetScript() == nil {
		return true, nil
	}
	binding := h.getBinding(src, a, &r.Meta().ElementMeta, &r.Meta().Type, nil)
	return h.EvalBool("source", binding, "process")
}

func (h *Handler) getBinding(src ocm.ComponentVersionAccess, a ocm.AccessSpec, m *compdesc.ElementMeta, typ *string, tgt ocm.Repository) map[string]interface{} {
	binding := map[string]interface{}{}
	if src != nil {
		binding["component"] = getCVAttrs(src)
	}

	if a != nil {
		binding["access"] = getData(a)
	}
	if m != nil {
		binding["element"] = getData(m)
	}
	if typ != nil {
		binding["element"].(map[string]interface{})["type"] = *typ
	}
	if tgt != nil {
		binding["target"] = getData(tgt.GetSpecification())
	}
	return binding
}

func getData(in interface{}) interface{} {
	var v interface{}

	d, err := json.Marshal(in)
	if err != nil {
		logrus.Error(err)
	}

	if err := json.Unmarshal(d, &v); err != nil {
		logrus.Error(err)
	}

	return v
}

func getCVAttrs(cv ocm.ComponentVersionAccess) map[string]interface{} {
	provider := map[string]interface{}{}
	data, err := json.Marshal(cv.GetDescriptor().Provider)
	if err != nil {
		logrus.Error(err)
	}
	json.Unmarshal(data, &provider)

	labels := cv.GetDescriptor().Labels.AsMap()

	values := map[string]interface{}{}
	values["name"] = cv.GetName()
	values["version"] = cv.GetVersion()
	values["provider"] = provider
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

	valueMap, ok := r.Value().(map[string]spiffing.Node)
	if !ok {
		return false, nil, nil, errors.ErrUnknown("transfer script field", key)
	}
	r = valueMap[key]
	if r == nil {
		return false, nil, nil, errors.ErrUnknown("transfer script field", key)
	}

	b, err := h.evalBoolValue(r)
	if err == nil {
		// flat boolean without result structure
		return b, nil, nil, nil
	}
	valueMap, ok = r.Value().(map[string]spiffing.Node)
	if !ok {
		return false, nil, nil, errors.ErrInvalid("transfer script result field type", dynaml.ExpressionType(r))
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
	v := valueMap["script"]
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
	v = valueMap["repospec"]
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
