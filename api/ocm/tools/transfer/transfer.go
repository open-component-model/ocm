package transfer

import (
	"context"
	"fmt"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/finalizer"
	"github.com/mandelsoft/logging"

	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/compdesc"
	ocmcpi "ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/none"
	"ocm.software/ocm/api/ocm/extensions/attrs/maxworkersattr"
	cpi "ocm.software/ocm/api/ocm/internal"
	"ocm.software/ocm/api/ocm/tools/transfer/internal"
	"ocm.software/ocm/api/ocm/tools/transfer/transferhandler/standard"
	"ocm.software/ocm/api/utils/errkind"
	common "ocm.software/ocm/api/utils/misc"
	runtimeutil "ocm.software/ocm/api/utils/runtime"
)

type WalkingState = common.WalkingState[*struct{}, interface{}]

type TransportClosure = common.NameVersionInfo[*struct{}]

func TransferVersion(printer common.Printer, closure TransportClosure, src ocmcpi.ComponentVersionAccess, tgt ocmcpi.Repository, handler TransferHandler) error {
	return TransferVersionWithContext(common.WithPrinter(context.Background(), common.AssurePrinter(printer)), closure, src, tgt, handler)
}

func TransferVersionWithContext(ctx context.Context, closure TransportClosure, src ocmcpi.ComponentVersionAccess, tgt ocmcpi.Repository, handler TransferHandler) error {
	if closure == nil {
		closure = TransportClosure{}
	}
	state := WalkingState{Closure: closure}
	return transferVersion(ctx, Logger(src), state, src, tgt, handler)
}

func transferVersion(ctx context.Context, log logging.Logger, state WalkingState, src ocmcpi.ComponentVersionAccess, tgt ocmcpi.Repository, handler TransferHandler) (rerr error) {
	printer := common.GetPrinter(ctx)
	if err := common.IsContextCanceled(ctx); err != nil {
		printer.Printf("transfer cancelled by caller\n")
		return err
	}
	nv := common.VersionedElementKey(src)
	log = log.WithValues("history", state.History.String(), "version", nv)
	if ok, err := state.Add(ocm.KIND_COMPONENTVERSION, nv); !ok {
		return err
	}
	log.Info("transferring version")
	printer.Printf("transferring version %q...\n", nv)
	if handler == nil {
		var err error
		handler, err = standard.New(standard.Overwrite())
		if err != nil {
			return err
		}
	}

	var finalize finalizer.Finalizer
	defer finalize.FinalizeWithErrorPropagation(&rerr)

	d := src.GetDescriptor()

	comp, err := tgt.LookupComponent(src.GetName())
	if err != nil {
		return errors.Wrapf(err, "%s: lookup target component", state.History)
	}
	finalize.Close(comp, "closing target component")

	var ok bool
	t, err := comp.LookupVersion(src.GetVersion())
	finalize.Close(t, "existing target version")

	// references have always to be handled, because of potentially different
	// transport modes, which could affect the desired access methods in
	// the target environment.

	// doTransport controls, whether the transport of the local component
	// version has to be re-considered.
	doTransport := true

	// doMerge controls. whether a potential current version in the target
	// environment has to be merged into the transported one.
	doMerge := false

	// doCopy controls, whether the artifact content has to be considered.
	doCopy := true

	if err != nil {
		if errors.IsErrNotFound(err) {
			t, err = comp.NewVersion(src.GetVersion())
			finalize.Close(t, "new target version")
		}
	} else {
		ok, err = handler.EnforceTransport(src, t)
		if err != nil {
			return err
		}
		if ok {
			//  execute transport as if the component version were not present
			//  on the target side.
		} else {
			// determine transport mode for component version present
			// on the target side.
			if eq := d.Equivalent(t.GetDescriptor()); eq.IsHashEqual() {
				if eq.IsEquivalent() {
					if !needsResourceTransport(src, d, t.GetDescriptor(), handler) {
						printer.Printf("  version %q already present -> skip transport\n", nv)
						doTransport = false
					} else {
						printer.Printf("  version %q already present -> but requires resource transport\n", nv)
					}
				} else {
					ok, err = handler.UpdateVersion(src, t)
					if err != nil {
						return err
					}
					if !ok {
						printer.Printf("  version %q requires update of volatile data, but skipped\n", nv)
						return nil
					}
					ok, err = handler.OverwriteVersion(src, t)
					if ok {
						printer.Printf("  warning: version %q already present, but transport enforced by overwrite option)\n", nv)
						doMerge = false
						doCopy = true
					} else {
						printer.Printf("  updating volatile properties of %q\n", nv)
						doMerge = true
						doCopy = false
					}
				}
			} else {
				msg := "  version %q already present, but"
				if eq.IsLocalHashEqual() {
					if eq.IsArtifactDetectable() {
						msg += " differs because some artifact digests are changed"
					} else {
						msg += " might differ, because not all artifact digests are known"
					}
				} else {
					if eq.IsArtifactDetectable() {
						if eq.IsArtifactEqual() {
							msg += " differs because signature relevant properties have been changed"
						} else {
							msg += " differs because some artifacts and signature relevant properties have been changed"
						}
					} else {
						msg += "differs because signature relevant properties have been changed (and not all artifact digests are known)"
					}
				}
				ok, err = handler.OverwriteVersion(src, t)
				if ok {
					doMerge = false
					printer.Printf("warning: "+msg+" (transport enforced by overwrite option)\n", nv)
				} else {
					printer.Printf(msg+" -> transport aborted (use option overwrite option to enforce transport)\n", nv)
					return errors.ErrAlreadyExists(ocm.KIND_COMPONENTVERSION, nv.String())
				}
			}
		}
	}
	if err != nil {
		return errors.Wrapf(err, "%s: creating target version", state.History)
	}

	list := errors.ErrListf("component references for %s", nv)
	if err := transferReferences(ctx, log, state, src, tgt, handler, d); err != nil {
		return err
	}

	if doTransport {
		var n *compdesc.ComponentDescriptor
		if doMerge {
			log.WithValues("source", src.GetDescriptor(), "target", t.GetDescriptor()).Info("  applying 2-way merge")
			n, err = internal.PrepareDescriptor(log, src.GetContext(), src.GetDescriptor(), t.GetDescriptor())
			if err != nil {
				return err
			}
		} else {
			n = src.GetDescriptor().Copy()
		}

		if !ocm.IsIntermediate(tgt.GetSpecification()) {
			if unstr, specErr := runtimeutil.ToUnstructuredTypedObject(tgt.GetSpecification()); specErr != nil {
				// Log the error as a warning, but don't fail the transfer.
				// The `log` variable from the function signature is suitable here.
				log.Warn("Failed to convert target repository specification to unstructured object for RepositoryContext",
					"error", specErr.Error(), // Use .Error() for string representation
					"spec_type", tgt.GetSpecification().GetType(), // Log the type of spec
				)
				// unstr remains nil, so it won't be appended.
			} else {
				// Only append if conversion was successful
				n.RepositoryContexts = append(n.RepositoryContexts, unstr)
			}
		}

		if !doMerge || doCopy {
			maxWorkers, err := maxworkersattr.Get(src.GetContext())
			if err != nil {
				return fmt.Errorf("failed to get max workers attribute: %w", err)
			}
			err = copyVersion(ctx, printer, log, state.History, src, t, n, handler, maxWorkers)
			if err != nil {
				return err
			}
		} else {
			*t.GetDescriptor() = *n
		}

		printer.Printf("...adding component version...\n")
		log.Info("  adding component version")
		list.Add(comp.AddVersion(t))
	}
	return list.Result()
}

func transferReferences(ctx context.Context, log logging.Logger, state WalkingState, src ocmcpi.ComponentVersionAccess, tgt ocmcpi.Repository, handler TransferHandler, d *compdesc.ComponentDescriptor) error {
	maxWorkers, err := maxworkersattr.Get(src.GetContext())
	if err != nil {
		return fmt.Errorf("failed to get max workers attribute: %w", err)
	}
	if len(d.References) > 0 {
		switch maxWorkers {
		case maxworkersattr.SingleWorker:
			for _, ref := range d.References {
				if err := transferReference(ctx, log, state, src, tgt, handler, ref); err != nil {
					return err
				}
			}
		default:
			if err := runWorkerPool(ctx, d.References, maxWorkers, func(ctx context.Context, ref compdesc.Reference) error {
				return transferReference(ctx, log, state, src, tgt, handler, ref)
			}); err != nil {
				return err
			}
		}
	}
	return nil
}

func transferReference(ctx context.Context, log logging.Logger, state WalkingState, src ocmcpi.ComponentVersionAccess, tgt ocmcpi.Repository, handler TransferHandler, ref compdesc.Reference) error {
	cv, shdlr, err := handler.TransferVersion(src.Repository(), src, &ref, tgt)
	if err != nil {
		return errors.Wrapf(err, "%s: nested component %s[%s:%s]",
			state.History, ref.GetName(), ref.ComponentName, ref.GetVersion())
	}
	if cv != nil {
		defer cv.Close()
		if err := transferVersion(common.AddPrinterGap(ctx, "  "),
			log.WithValues("ref", ref.Name), state, cv, tgt, shdlr); err != nil {
			return errors.Wrapf(err, "%s: transferring reference %s[%s:%s]",
				state.History, ref.GetName(), ref.ComponentName, ref.GetVersion())
		}
	}
	return nil
}

func CopyVersion(printer common.Printer, log logging.Logger, hist common.History, src ocm.ComponentVersionAccess, t ocm.ComponentVersionAccess, handler TransferHandler) (rerr error) {
	return CopyVersionWithContext(context.Background(), printer, log, hist, src, t, handler)
}

func CopyVersionWithContext(cctx context.Context, printer common.Printer, log logging.Logger, hist common.History, src ocm.ComponentVersionAccess, t ocm.ComponentVersionAccess, handler TransferHandler) (rerr error) {
	maxWorkers, err := maxworkersattr.Get(src.GetContext())
	if err != nil {
		return fmt.Errorf("failed to get max workers attribute: %w", err)
	}

	return copyVersion(cctx, printer, log, hist, src, t, src.GetDescriptor().Copy(), handler, maxWorkers)
}

func notifyArtifactInfo(printer common.Printer, log logging.Logger, kind string, index int, meta compdesc.ArtifactMetaAccess, hint string, msgs ...interface{}) {
	msg := "copying"
	cmsg := "..."
	if len(msgs) > 0 {
		if m, ok := msgs[0].(string); ok {
			msg = fmt.Sprintf(m, msgs[1:]...)
		} else {
			msg = fmt.Sprint(msgs...)
		}
		cmsg = " (" + msg + ")"
	}
	if printer != nil {
		if hint != "" {
			printer.Printf("...%s %d %s[%s](%s)%s\n", kind, index, meta.GetName(), meta.GetType(), hint, cmsg)
		} else {
			printer.Printf("...%s %d %s[%s]%s\n", kind, index, meta.GetName(), meta.GetType(), cmsg)
		}
	}
	if hint != "" {
		log.Debug("handle artifact", "kind", kind, "name", meta.GetName(), "type", meta.GetType(), "index", index, "hint", hint, "message", msg)
	} else {
		log.Debug("handle artifact", "kind", kind, "name", meta.GetName(), "type", meta.GetType(), "index", index, "message", msg)
	}
}

func copyVersion(
	ctx context.Context,
	printer common.Printer,
	log logging.Logger,
	hist common.History,
	src ocm.ComponentVersionAccess,
	t ocm.ComponentVersionAccess,
	prep *compdesc.ComponentDescriptor,
	handler TransferHandler,
	workers uint,
) (rerr error) {
	var finalize finalizer.Finalizer
	defer errors.PropagateError(&rerr, finalize.Finalize)

	if handler == nil {
		handler = standard.NewDefaultHandler(nil)
	}

	srccd := src.GetDescriptor()
	cur := *t.GetDescriptor()
	*t.GetDescriptor() = *prep

	switch workers {
	case maxworkersattr.SingleWorker:
		log.Debug("single worker environment detected, using sequential copy")
		return copyVersionSequentially(ctx, src, &finalize, hist, handler, &cur, srccd, printer, log, t)

	default:
		log.Debug("concurrent worker environment detected, using concurrent copy", "workers", workers)
		return copyVersionConcurrently(ctx, printer, log, hist, src, t, workers, &finalize, handler, &cur, srccd)
	}
}

func copyVersionSequentially(ctx context.Context, src ocm.ComponentVersionAccess, finalize *finalizer.Finalizer, hist common.History, handler TransferHandler, cur *compdesc.ComponentDescriptor, srccd *compdesc.ComponentDescriptor, printer common.Printer, log logging.Logger, t ocm.ComponentVersionAccess) error {
	for i, r := range src.GetResources() {
		if err := copyResource(src, finalize, hist, handler, cur, srccd, printer, log, t, r, i); err != nil {
			return err
		}
	}

	for i, r := range src.GetSources() {
		if err := copySource(ctx, src, hist, r, handler, printer, log, i, t); err != nil {
			return err
		}
	}
	return nil
}

func copyVersionConcurrently(
	ctx context.Context,
	printer common.Printer,
	log logging.Logger,
	hist common.History,
	src ocm.ComponentVersionAccess,
	target ocm.ComponentVersionAccess,
	maxWorkers uint,
	finalize *finalizer.Finalizer,
	handler TransferHandler,
	curDesc, srcDesc *compdesc.ComponentDescriptor,
) error {
	type transferTask struct {
		id   string
		exec func(ctx context.Context) error
	}

	var tasks []transferTask

	// Prepare all tasks first (no side effects yet)
	for i, r := range src.GetResources() {
		tasks = append(tasks, transferTask{
			id: fmt.Sprintf("resource-%d", i),
			exec: func(ctx context.Context) error {
				return copyResource(src, finalize, hist, handler, curDesc, srcDesc, printer, log, target, r, i)
			},
		})
	}
	for i, s := range src.GetSources() {
		tasks = append(tasks, transferTask{
			id: fmt.Sprintf("source-%d", i),
			exec: func(ctx context.Context) error {
				return copySource(ctx, src, hist, s, handler, printer, log, i, target)
			},
		})
	}

	// Run all tasks using the generic worker pool
	return runWorkerPool(ctx, tasks, maxWorkers, func(ctx context.Context, t transferTask) error {
		log.Debug("starting transfer task", "task", t.id)
		if err := t.exec(ctx); err != nil {
			return fmt.Errorf("%s failed: %w", t.id, err)
		}
		return nil
	})
}

func copySource(cctx context.Context, src ocm.ComponentVersionAccess, hist common.History, srcAccess cpi.SourceAccess, handler TransferHandler, printer common.Printer, log logging.Logger, i int, t ocm.ComponentVersionAccess) error {
	var m ocmcpi.AccessMethod

	if err := common.IsContextCanceled(cctx); err != nil {
		printer.Printf("cancelled by caller\n")
		return err
	}

	a, err := srcAccess.Access()
	if err == nil {
		m, err = a.AccessMethod(src)
	}
	if err == nil {
		ok := a.IsLocal(src.GetContext())
		if !ok {
			if !none.IsNone(a.GetKind()) {
				ok, err = handler.TransferSource(src, a, srcAccess)
				if err == nil && !ok {
					log.Info("transport omitted", "source", srcAccess.Meta().Name, "index", i, "access", a.GetType())
				}
			}
		}
		if ok {
			// sources do not have digests so far, so they have to copied, always.
			hint := ocmcpi.ArtifactNameHint(a, src)
			notifyArtifactInfo(printer, log, "source", i, srcAccess.Meta(), hint)
			err = errors.Join(err, handler.HandleTransferSource(srcAccess, m, hint, t))
		}
		err = errors.Join(err, m.Close())
	}
	if err != nil {
		if !errors.IsErrUnknownKind(err, errkind.KIND_ACCESSMETHOD) {
			return errors.Wrapf(err, "%s: transferring source %d", hist, i)
		}
		printer.Printf("WARN: %s: transferring source %d: %s (enforce transport by reference)\n", hist, i, err)
	}
	return nil
}

func copyResource(src ocm.ComponentVersionAccess, finalize *finalizer.Finalizer, hist common.History, handler TransferHandler, currentDesc, sourceDesc *compdesc.ComponentDescriptor, printer common.Printer, log logging.Logger, t ocm.ComponentVersionAccess, r cpi.ResourceAccess, i int) error {
	nested := finalize.Nested()
	a, err := r.Access()
	if err != nil {
		return err
	}
	m, err := a.AccessMethod(src)
	nested.Close(m, fmt.Sprintf("%s: transferring resource %d: closing access method", hist, i))
	if err != nil {
		return err
	}

	shouldTransfer := a.IsLocal(src.GetContext())
	if !shouldTransfer && !none.IsNone(a.GetKind()) {
		if shouldTransfer, err = handler.TransferResource(src, a, r); err != nil || !shouldTransfer {
			return err
		}
	}
	if !shouldTransfer {
		return nested.Finalize()
	}

	hint := ocmcpi.ArtifactNameHint(a, src)
	old, err := currentDesc.GetResourceByIdentity(r.Meta().GetIdentity(sourceDesc.Resources))

	changed := err != nil || old.Digest == nil || !old.Digest.Equal(r.Meta().Digest)
	needsTransportByValue := err == nil && needsTransport(src.GetContext(), r, &old)

	if changed || needsTransportByValue {
		var msgs []interface{}
		if !errors.IsErrNotFound(err) {
			if err != nil {
				return err
			}
			if !changed {
				msgs = append(msgs, "copy")
			} else {
				msgs = append(msgs, "overwrite")
			}
		}
		notifyArtifactInfo(printer, log, "resource", i, r.Meta(), hint, msgs...)
		return handler.HandleTransferResource(r, m, hint, t)
	}

	if err := t.SetResource(r.Meta(), old.Access, ocm.ModifyElement(), ocm.SkipVerify(), ocm.DisableExtraIdentityDefaulting()); err != nil {
		return fmt.Errorf("failed to set resource based on existing access method %d: %w", i, err)
	}
	notifyArtifactInfo(printer, log, "resource", i, r.Meta(), hint, "already present")
	return nil
}
