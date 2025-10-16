package standard

import (
	"time"

	"github.com/mandelsoft/goutils/errors"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/compdesc"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/cpi/accspeccpi"
	"ocm.software/ocm/api/ocm/resolvers"
	"ocm.software/ocm/api/ocm/tools/transfer/transferhandler"
	"ocm.software/ocm/api/utils/accessio"
)

type Handler struct {
	opts *Options
}

func NewDefaultHandler(opts *Options) *Handler {
	if opts == nil {
		opts = &Options{}
	}
	return &Handler{opts: opts}
}

func New(opts ...transferhandler.TransferOption) (transferhandler.TransferHandler, error) {
	defaultOpts := &Options{}
	err := transferhandler.ApplyOptions(defaultOpts, opts...)
	if err != nil {
		return nil, err
	}
	return NewDefaultHandler(defaultOpts), nil
}

func (h *Handler) UpdateVersion(src ocm.ComponentVersionAccess, tgt ocm.ComponentVersionAccess) (bool, error) {
	return !h.opts.IsSkipUpdate(), nil
}

func (h *Handler) EnforceTransport(src ocm.ComponentVersionAccess, tgt ocm.ComponentVersionAccess) (bool, error) {
	return h.opts.IsTransportEnforced(), nil
}

func (h *Handler) OverwriteVersion(src ocm.ComponentVersionAccess, tgt ocm.ComponentVersionAccess) (bool, error) {
	return h.opts.IsOverwrite(), nil
}

func (h *Handler) TransferVersion(repo ocm.Repository, src ocm.ComponentVersionAccess, meta *compdesc.Reference, tgt ocm.Repository) (ocm.ComponentVersionAccess, transferhandler.TransferHandler, error) {
	if src == nil || h.opts.IsRecursive() {
		if h.opts.IsStopOnExistingVersion() && tgt != nil {
			if found, err := tgt.ExistsComponentVersion(meta.ComponentName, meta.Version); found || err != nil {
				return nil, nil, errors.Wrapf(err, "failed looking up in target")
			}
		}
		compoundResolver := resolvers.NewCompoundResolver(repo, h.opts.GetResolver())
		cv, err := compoundResolver.LookupComponentVersion(meta.GetComponentName(), meta.Version)
		return cv, h, err
	}
	return nil, nil, nil
}

func (h *Handler) TransferResource(src ocm.ComponentVersionAccess, a ocm.AccessSpec, r ocm.ResourceAccess) (bool, error) {
	if h.opts.IsAccessTypeOmitted(a.GetType()) {
		return false, nil
	}
	if h.opts.IsLocalResourcesByValue() {
		if r.Meta().Relation == metav1.LocalRelation {
			return true, nil
		}
	}
	return h.opts.IsResourcesByValue(), nil
}

func (h *Handler) TransferSource(src ocm.ComponentVersionAccess, a ocm.AccessSpec, r ocm.SourceAccess) (bool, error) {
	if h.opts.IsAccessTypeOmitted(a.GetType()) {
		return false, nil
	}
	return h.opts.IsSourcesByValue(), nil
}

func (h *Handler) HandleTransferResource(r ocm.ResourceAccess, m cpi.AccessMethod, hint string, t ocm.ComponentVersionAccess) error {
	blob, err := accspeccpi.BlobAccessForAccessMethod(m)
	if err != nil {
		return err
	}
	defer blob.Close()
	return accessio.Retry(h.opts.GetRetries(), time.Second, func() error {
		return t.SetResourceBlob(r.Meta(), blob, hint, h.GlobalAccess(t.GetContext(), m), ocm.SkipVerify(), ocm.DisableExtraIdentityDefaulting())
	})
}

func (h *Handler) HandleTransferSource(r ocm.SourceAccess, m cpi.AccessMethod, hint string, t ocm.ComponentVersionAccess) error {
	blob, err := accspeccpi.BlobAccessForAccessMethod(m)
	if err != nil {
		return err
	}
	defer blob.Close()
	return accessio.Retry(h.opts.GetRetries(), time.Second, func() error {
		return t.SetSourceBlob(r.Meta(), blob, hint, h.GlobalAccess(t.GetContext(), m), ocm.DisableExtraIdentityDefaulting())
	})
}

func (h *Handler) GlobalAccess(ctx ocm.Context, m ocm.AccessMethod) ocm.AccessSpec {
	if h.opts.IsKeepGlobalAccess() {
		return m.AccessSpec().GlobalAccessSpec(ctx)
	}
	return nil
}
