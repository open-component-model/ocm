package plugin

import (
	"encoding/json"
	"sync"

	"github.com/mandelsoft/goutils/errors"
	"golang.org/x/exp/slices"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/compdesc"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	v2 "ocm.software/ocm/api/ocm/compdesc/versions/v2"
	"ocm.software/ocm/api/ocm/extensions/attrs/plugincacheattr"
	"ocm.software/ocm/api/ocm/plugin"
	"ocm.software/ocm/api/ocm/plugin/descriptor"
	"ocm.software/ocm/api/ocm/plugin/ppi"
	"ocm.software/ocm/api/ocm/tools/transfer/transferhandler"
	"ocm.software/ocm/api/ocm/tools/transfer/transferhandler/standard"
	"ocm.software/ocm/api/utils/runtime"
)

type Handler struct {
	lock sync.Mutex
	standard.Handler
	opts   *Options
	plugin plugin.Plugin
	desc   *descriptor.TransferHandlerDescriptor
}

func New(opts ...transferhandler.TransferOption) (transferhandler.TransferHandler, error) {
	options := &Options{}
	err := transferhandler.ApplyOptions(options, opts...)
	if err != nil {
		return nil, err
	}

	return &Handler{
		Handler: *standard.NewDefaultHandler(&options.Options),
		opts:    options,
	}, nil
}

func (h *Handler) GetConfig() []byte {
	return h.opts.GetConfig()
}

func (h *Handler) ResolvePlugin(ctx ocm.ContextProvider) (plugin.Plugin, error) {
	h.lock.Lock()
	defer h.lock.Unlock()

	if h.plugin == nil {
		pi := plugincacheattr.Get(ctx)
		if pi == nil {
			return nil, errors.ErrUnknown(plugin.KIND_PLUGIN, h.opts.plugin)
		}
		p := pi.Get(h.opts.plugin)
		if p == nil {
			return nil, errors.ErrUnknown(plugin.KIND_PLUGIN, h.opts.plugin)
		}
		h.plugin = p
		h.desc = p.GetDescriptor().TransferHandlers.Get(h.opts.GetTransferHandler())
		if h.desc == nil {
			return nil, errors.ErrUnknown(plugin.KIND_TRANSFERHANDLER, h.opts.plugin)
		}
	}
	return h.plugin, nil
}

func filterLabels(in metav1.Labels, filter *[]string) metav1.Labels {
	var r metav1.Labels

	if filter == nil {
		return in.Copy()
	}
	if len(*filter) > 0 {
		for _, l := range in {
			if filter == nil || slices.Contains(*filter, l.Name) {
				r = append(r, l)
			}
		}
	}
	return r
}

func (h *Handler) transferOptions() (*plugin.TransferOptions, error) {
	var special *json.RawMessage

	if h.opts.config != nil {
		s, err := runtime.ToJSON(h.opts.config)
		if err != nil {
			return nil, err
		}
		special = &s
	}
	return &plugin.TransferOptions{
		Recursive:         h.opts.GetRecursive(),
		ResourcesByValue:  h.opts.GetResourcesByValue(),
		LoalByValue:       h.opts.GetLocalResourcesByValue(),
		SourcesByValue:    h.opts.GetSourcesByValue(),
		KeepGlobalAccess:  h.opts.GetKeepGlobalAccess(),
		StopOnExisting:    h.opts.GetStopOnExistingVersion(),
		EnforceTransport:  h.opts.GetEnforceTransport(),
		Overwrite:         h.opts.GetOverwrite(),
		SkipUpdate:        h.opts.GetSkipUpdate(),
		OmitAccessTypes:   h.opts.GetOmittedAccessTypes(),
		OmitArtifactTypes: h.opts.GetOmittedArtifactTypes(),
		Special:           special,
	}, nil
}

func (h *Handler) sourceComponentVersion(src ocm.ComponentVersionAccess, question string) (*plugin.SourceComponentVersion, error) {
	repo := src.Repository()
	defer repo.Close()

	repospec, err := ocm.ToGenericRepositorySpec(repo.GetSpecification())
	if err != nil {
		return nil, err
	}

	filter := h.desc.GetQuestion(question).Labels

	return &plugin.SourceComponentVersion{
		Name:    src.GetName(),
		Version: src.GetVersion(),
		Provider: metav1.Provider{
			Name:   src.GetProvider().GetName(),
			Labels: filterLabels(src.GetProvider().Labels, filter),
		},
		Repository: *repospec,
		Labels:     filterLabels(src.GetDescriptor().Labels, filter),
	}, nil
}

func (h *Handler) respositoryTarget(tgt ocm.ComponentVersionAccess) (*plugin.TargetRepositorySpec, error) {
	repo := tgt.Repository()
	defer repo.Close()

	return ocm.ToGenericRepositorySpec(repo.GetSpecification())
}

func (h *Handler) artifact(ctx ocm.Context, art ocm.AccessSpec, ra *compdesc.ElementMeta, question string) (*plugin.Artifact, error) {
	filter := h.desc.GetQuestion(question).Labels

	a := art.Info(ctx)
	s, err := ocm.ToGenericAccessSpec(art)
	if err != nil {
		return nil, err
	}

	return &plugin.Artifact{
		Meta: v2.ElementMeta{
			Name:          ra.Name,
			Version:       ra.Version,
			ExtraIdentity: ra.ExtraIdentity.Copy(),
			Labels:        filterLabels(ra.Labels, filter),
		},
		Access:     *s,
		AccessInfo: *a,
	}, nil
}

func (h *Handler) askComponentVersionQuestion(question string, src ocm.ComponentVersionAccess, tgt ocm.ComponentVersionAccess) (bool, error) {
	p, err := h.ResolvePlugin(src.GetContext())
	if err != nil {
		return false, err
	}

	s, err := h.sourceComponentVersion(src, question)
	if err != nil {
		return false, err
	}
	t, err := h.respositoryTarget(tgt)
	if err != nil {
		return false, err
	}
	o, err := h.transferOptions()
	if err != nil {
		return false, err
	}

	args := plugin.ComponentVersionQuestion{
		Source:  *s,
		Target:  *t,
		Options: *o,
	}

	r, err := p.AskTransferQuestion(h.opts.handler, question, args)
	if err != nil {
		return false, err
	}
	return r.Decision, nil
}

func (h *Handler) askArtifactQuestion(question string, src ocm.ComponentVersionAccess, art ocm.AccessSpec, ra *compdesc.ElementMeta) (bool, error) {
	p, err := h.ResolvePlugin(src.GetContext())
	if err != nil {
		return false, err
	}

	s, err := h.sourceComponentVersion(src, question)
	if err != nil {
		return false, err
	}
	a, err := h.artifact(src.GetContext(), art, ra, question)
	if err != nil {
		return false, err
	}
	o, err := h.transferOptions()
	if err != nil {
		return false, err
	}

	args := plugin.ArtifactQuestion{
		Source:   *s,
		Artifact: *a,
		Options:  *o,
	}

	r, err := p.AskTransferQuestion(h.opts.handler, question, args)
	if err != nil {
		return false, err
	}
	return r.Decision, nil
}

func (h *Handler) UpdateVersion(src ocm.ComponentVersionAccess, tgt ocm.ComponentVersionAccess) (bool, error) {
	_, err := h.ResolvePlugin(src.GetContext())
	if err != nil {
		return false, err
	}
	if h.desc.GetQuestion(plugin.Q_UPDATE_VERSION) == nil {
		return h.Handler.UpdateVersion(src, tgt)
	}
	return h.askComponentVersionQuestion(plugin.Q_UPDATE_VERSION, src, tgt)
}

func (h *Handler) EnforceTransport(src ocm.ComponentVersionAccess, tgt ocm.ComponentVersionAccess) (bool, error) {
	_, err := h.ResolvePlugin(src.GetContext())
	if err != nil {
		return false, err
	}
	if h.desc.GetQuestion(plugin.Q_ENFORCE_TRANSPORT) == nil {
		return h.Handler.EnforceTransport(src, tgt)
	}
	return h.askComponentVersionQuestion(plugin.Q_ENFORCE_TRANSPORT, src, tgt)
}

func (h *Handler) OverwriteVersion(src ocm.ComponentVersionAccess, tgt ocm.ComponentVersionAccess) (bool, error) {
	_, err := h.ResolvePlugin(src.GetContext())
	if err != nil {
		return false, err
	}
	if h.desc.GetQuestion(plugin.Q_OVERWRITE_VERSION) == nil {
		return h.Handler.OverwriteVersion(src, tgt)
	}
	return h.askComponentVersionQuestion(plugin.Q_OVERWRITE_VERSION, src, tgt)
}

func (h *Handler) TransferVersion(repo ocm.Repository, src ocm.ComponentVersionAccess, meta *compdesc.Reference, tgt ocm.Repository) (ocm.ComponentVersionAccess, transferhandler.TransferHandler, error) {
	_, err := h.ResolvePlugin(src.GetContext())
	if err != nil {
		return nil, nil, err
	}
	if h.desc.GetQuestion(plugin.Q_TRANSFER_VERSION) == nil {
		return h.Handler.TransferVersion(repo, src, meta, tgt)
	}

	filter := h.desc.GetQuestion(plugin.Q_TRANSFER_VERSION).Labels
	p, err := h.ResolvePlugin(src.GetContext())
	if err != nil {
		return nil, nil, err
	}

	s, err := h.sourceComponentVersion(src, plugin.Q_TRANSFER_VERSION)
	if err != nil {
		return nil, nil, err
	}

	t, err := ocm.ToGenericRepositorySpec(repo.GetSpecification())
	if err != nil {
		return nil, nil, err
	}
	o, err := h.transferOptions()
	if err != nil {
		return nil, nil, err
	}

	args := plugin.ComponentReferenceQuestion{
		Source: *s,
		Target: *t,
		ElementMeta: v2.ElementMeta{
			Name:          meta.Name,
			Version:       meta.Version,
			ExtraIdentity: meta.ExtraIdentity.Copy(),
			Labels:        filterLabels(meta.Labels, filter),
		},
		Options: *o,
	}

	r, err := p.AskTransferQuestion(h.opts.handler, plugin.Q_TRANSFER_VERSION, args)
	if err != nil {
		return nil, nil, err
	}

	// TODO: evaluate transfer handler and repo
	_ = r
	return h.Handler.TransferVersion(repo, src, meta, tgt)
}

func (h *Handler) TransferResource(src ocm.ComponentVersionAccess, a ocm.AccessSpec, r ocm.ResourceAccess) (bool, error) {
	_, err := h.ResolvePlugin(src.GetContext())
	if err != nil {
		return false, err
	}
	if h.desc.GetQuestion(plugin.Q_TRANSFER_RESOURCE) == nil {
		return h.Handler.TransferResource(src, a, r)
	}
	return h.askArtifactQuestion(ppi.Q_TRANSFER_RESOURCE, src, a, &r.Meta().ElementMeta)
}

func (h *Handler) TransferSource(src ocm.ComponentVersionAccess, a ocm.AccessSpec, r ocm.SourceAccess) (bool, error) {
	_, err := h.ResolvePlugin(src.GetContext())
	if err != nil {
		return false, err
	}
	if h.desc.GetQuestion(plugin.Q_TRANSFER_SOURCE) == nil {
		return h.Handler.TransferSource(src, a, r)
	}
	return h.askArtifactQuestion(ppi.Q_TRANSFER_SOURCE, src, a, &r.Meta().ElementMeta)
}
