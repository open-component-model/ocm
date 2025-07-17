package transfer

import (
	"context"
	"fmt"
	"runtime"
	"sync"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/finalizer"
	"github.com/mandelsoft/logging"

	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/compdesc"
	ocmcpi "ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/none"
	"ocm.software/ocm/api/ocm/extensions/attrs/maxworkersattr"
	"ocm.software/ocm/api/ocm/tools/transfer/internal"
	"ocm.software/ocm/api/ocm/tools/transfer/transferhandler/standard"
	common "ocm.software/ocm/api/utils/misc"
	runtimeutil "ocm.software/ocm/api/utils/runtime" 
)

type WalkingState = common.WalkingState[*struct{}, interface{}]

type TransportClosure = common.NameVersionInfo[*struct{}]

// TransferWorkersEnvVar is the environment variable to configure the number of transfer workers.
const TransferWorkersEnvVar = "OCM_TRANSFER_WORKERS" // This constant is now technically unused, but kept for context.

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

	var wg sync.WaitGroup
	var mu sync.Mutex
	list := errors.ErrListf("component references for %s", nv)

	for _, r := range d.References {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cv, shdlr, err := handler.TransferVersion(src.Repository(), src, &r, tgt)
			if err != nil {
				mu.Lock()
				list.Add(errors.Wrapf(err, "%s: nested component %s[%s:%s]", state.History, r.GetName(), r.ComponentName, r.GetVersion()))
				mu.Unlock()
				return
			}
			if cv != nil {
				err1 := transferVersion(ctx, log.WithValues("ref", r.Name), state, cv, tgt, shdlr)
				err2 := cv.Close()
				mu.Lock()
				list.Add(err1)
				list.Addf(nil, err2, "closing reference %s", r.Name)
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

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

		var unstr *runtimeutil.UnstructuredTypedObject
		if !ocm.IsIntermediate(tgt.GetSpecification()) {
			// Capture the error specifically for this operation
			specErr := error(nil) // Declare a local error variable
			unstr, specErr = runtimeutil.ToUnstructuredTypedObject(tgt.GetSpecification())
			if specErr != nil {
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
			numWorkers := calculateEffectiveTransferWorkers(ctx) 
			err = copyVersionWithWorkerPool(ctx, printer, log, state.History, src, t, n, handler, numWorkers)
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

func CopyVersion(printer common.Printer, log logging.Logger, hist common.History, src ocm.ComponentVersionAccess, t ocm.ComponentVersionAccess, handler TransferHandler) (rerr error) {
	return CopyVersionWithContext(context.Background(), printer, log, hist, src, t, handler)
}

func CopyVersionWithContext(cctx context.Context, printer common.Printer, log logging.Logger, hist common.History, src ocm.ComponentVersionAccess, t ocm.ComponentVersionAccess, handler TransferHandler) (rerr error) {
	numWorkers := calculateEffectiveTransferWorkers(cctx) 
	return copyVersionWithWorkerPool(cctx, printer, log, hist, src, t, src.GetDescriptor().Copy(), handler, numWorkers)
}

// calculateEffectiveTransferWorkers determines the number of workers to use.
// It prioritizes an explicit attribute setting, then falls back to CPU-based auto-detection.
func calculateEffectiveTransferWorkers(ctx context.Context) int {
	// First, obtain the OCM data context from the provided Go context.
	// `ocm.DefaultContext()` is a common way to get it if it's not directly `datacontext.Context`.
	ocmCtx := ocm.DefaultContext()

	// Get the workers value from the attribute.
	// This will be:
	// - User-defined value (if -X or .ocmconfig is used and > 0)
	// - 0 (if user explicitly set -X maxworkers=0, OR if the attribute was not set at all)
	attributeWorkers := maxworkersattr.Get(ocmCtx)

	// If attributeWorkers is 0, it means either user explicitly set 0, or attribute was not set (default to 0).
	// In both cases, this signals to use the CPU-based auto-detection.
	if attributeWorkers == 0 {
		return determineWorkersFromCPU()
	}

	// Otherwise (if attributeWorkers is > 0), use the user-defined value.
	return int(attributeWorkers)
}

// determineWorkersFromCPU implements your CPU-based logic.
func determineWorkersFromCPU() int {
	numCPU := runtime.NumCPU()

	if numCPU <= 2 {
		return 1
	} else if numCPU <= 4 {
		return 2
	} else if numCPU <= 8 {
		return 4
	} else {
		return numCPU / 2
	}
}

func copyVersionWithWorkerPool(ctx context.Context, printer common.Printer, log logging.Logger, hist common.History, src ocm.ComponentVersionAccess, t ocm.ComponentVersionAccess, prep *compdesc.ComponentDescriptor, handler TransferHandler, maxWorkers int) (rerr error) {
	type transferTask struct {
		task func() error
		id   string
	}

	var finalize finalizer.Finalizer
	defer errors.PropagateError(&rerr, finalize.Finalize)

	if handler == nil {
		handler = standard.NewDefaultHandler(nil)
	}

	srccd := src.GetDescriptor()
	cur := *t.GetDescriptor()
	*t.GetDescriptor() = *prep

	log.Info("  transferring resources and sources using worker pool", "workers", maxWorkers)
	taskBufferSize := maxWorkers * 2 // Example: buffer size is twice the number of workers
	if taskBufferSize == 0 {         // Handle case where maxWorkers might be 0 or 1, ensure at least a small buffer or 1
		taskBufferSize = 1
	}
	tasks := make(chan transferTask, taskBufferSize) // Now a buffered channel

	// The 'errChan' should also ideally be buffered to avoid blocking when reporting multiple errors.
	errChan := make(chan error, len(src.GetResources())+len(src.GetSources()))

	var wg sync.WaitGroup
	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for item := range tasks {
				if err := item.task(); err != nil {
					errChan <- err
				}
			}
		}(i)
	}

	// Helper function to handle resource transfer
	handleResourceTransfer := func(i int, r ocmcpi.ResourceAccess) func() error {
		return func() error {
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
			ok := a.IsLocal(src.GetContext())
			if !ok && !none.IsNone(a.GetKind()) {
				ok, err = handler.TransferResource(src, a, r)
				if err != nil || !ok {
					return err
				}
			}
			if ok {
				hint := ocmcpi.ArtifactNameHint(a, src)
				old, err := cur.GetResourceByIdentity(r.Meta().GetIdentity(srccd.Resources))
				changed := err != nil || old.Digest == nil || !old.Digest.Equal(r.Meta().Digest)
				valueNeeded := err == nil && needsTransport(src.GetContext(), r, &old)
				if changed || valueNeeded {
					notifyArtifactInfo(printer, log, "resource", i, r.Meta(), hint, "copy")
					return handler.HandleTransferResource(r, m, hint, t)
				} else if err == nil {
					t.SetResource(r.Meta(), old.Access, ocm.ModifyElement(), ocm.SkipVerify(), ocm.DisableExtraIdentityDefaulting())
					notifyArtifactInfo(printer, log, "resource", i, r.Meta(), hint, "already present")
				}
			}
			return nested.Finalize()
		}
	}

	// Helper function to handle source transfer
	handleSourceTransfer := func(i int, r ocmcpi.SourceAccess) func() error {
		return func() error {
			a, err := r.Access()
			if err != nil {
				return err
			}
			m, err := a.AccessMethod(src)
			if err != nil {
				return err
			}
			ok := a.IsLocal(src.GetContext())
			if !ok && !none.IsNone(a.GetKind()) {
				ok, err = handler.TransferSource(src, a, r)
				if err != nil || !ok {
					return err
				}
			}
			if ok {
				hint := ocmcpi.ArtifactNameHint(a, src)
				notifyArtifactInfo(printer, log, "source", i, r.Meta(), hint)
				if err := handler.HandleTransferSource(r, m, hint, t); err != nil {
					return err
				}
			}
			return m.Close()
		}
	}

	go func() {
		for i, r := range src.GetResources() {
			tasks <- transferTask{
				id:   fmt.Sprintf("resource-%d", i),
				task: handleResourceTransfer(i, r),
			}
		}

		for i, r := range src.GetSources() {
			tasks <- transferTask{
				id:   fmt.Sprintf("source-%d", i),
				task: handleSourceTransfer(i, r),
			}
		}
		close(tasks)
	}()

	go func() {
		wg.Wait()
		close(errChan)
	}()

	errList := errors.ErrListf("transfer resources and sources")
	for e := range errChan {
		errList.Add(e)
	}
	return errList.Result()
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


