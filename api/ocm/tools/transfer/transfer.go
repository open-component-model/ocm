package transfer

import (
	"fmt"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/finalizer"
	"github.com/mandelsoft/logging"

	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/compdesc"
	ocmcpi "ocm.software/ocm/api/ocm/cpi"
	//"ocm.software/ocm/api/ocm/extensions/accessmethods/none"
	"ocm.software/ocm/api/ocm/tools/transfer/internal"
	"ocm.software/ocm/api/ocm/tools/transfer/transferhandler/standard"
	//"ocm.software/ocm/api/utils/errkind"
	common "ocm.software/ocm/api/utils/misc"
	"ocm.software/ocm/api/utils/runtime"

	//"sync"
	//"time"
    
    
)

type WalkingState = common.WalkingState[*struct{}, interface{}]

type TransportClosure = common.NameVersionInfo[*struct{}]

func TransferVersion(printer common.Printer, closure TransportClosure, src ocmcpi.ComponentVersionAccess, tgt ocmcpi.Repository, handler TransferHandler) error {
	if closure == nil {
		closure = TransportClosure{}
	}
	state := WalkingState{Closure: closure}
	return transferVersion(common.AssurePrinter(printer), Logger(src), state, src, tgt, handler)
}


type TransferTask struct {
    Resource ocm.ResourceAccess
    Source   ocm.SourceAccess
    Method   ocmcpi.AccessMethod
    Hint     string
    Target   ocm.ComponentVersionAccess
}


func createWorkerPool(numWorkers int, handler TransferHandler) (chan TransferTask, chan error) {
    tasks := make(chan TransferTask, numWorkers)
    results := make(chan error, numWorkers)

    for i := 0; i < numWorkers; i++ {
        go worker(i, tasks, results, handler)
    }

    return tasks, results
}

func worker(id int, tasks <-chan TransferTask, results chan<- error, handler TransferHandler) {
    fmt.Printf("Worker %d: Initialized\n", id)
    for task := range tasks {
		fmt.Printf("task")
        if task.Resource != nil {
            //fmt.Printf("Worker %d: Starting transfer of resource %s\n", id, task.Resource.Meta().Name)
            err := handler.HandleTransferResource(task.Resource, task.Method, task.Hint, task.Target)
            //fmt.Printf("Worker %d: Finished transfer of resource %s\n", id, task.Resource.Meta().Name)
            results <- err
        } else if task.Source != nil {
            //fmt.Printf("Worker %d: Starting transfer of source %s\n", id, task.Source.Meta().Name)
            err := handler.HandleTransferSource(task.Source, task.Method, task.Hint, task.Target)
            //fmt.Printf("Worker %d: Finished transfer of source %s\n", id, task.Source.Meta().Name)
            results <- err
        }
    }
}

func transferVersion(printer common.Printer, log logging.Logger, state WalkingState, src ocmcpi.ComponentVersionAccess, tgt ocmcpi.Repository, handler TransferHandler) (rerr error) {
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

    doTransport := true
    doMerge := false
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
            // Transport enforced
        } else {
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

    subp := printer.AddGap("  ")
    list := errors.ErrListf("component references for %s", nv)
    log.Info("  transferring references")
    for _, r := range d.References {
        cv, shdlr, err := handler.TransferVersion(src.Repository(), src, &r, tgt)
        if err != nil {
            return errors.Wrapf(err, "%s: nested component %s[%s:%s]", state.History, r.GetName(), r.ComponentName, r.GetVersion())
        }
        if cv != nil {
            list.Add(transferVersion(subp, log.WithValues("ref", r.Name), state, cv, tgt, shdlr))
            list.Addf(nil, cv.Close(), "closing reference %s", r.Name)
        }
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

        var unstr *runtime.UnstructuredTypedObject
        if !ocm.IsIntermediate(tgt.GetSpecification()) {
            unstr, err = runtime.ToUnstructuredTypedObject(tgt.GetSpecification())
            if err != nil {
                unstr = nil
            }
        }
        if unstr != nil {
            n.RepositoryContexts = append(n.RepositoryContexts, unstr)
        }

        if !doMerge || doCopy {
            numWorkers := 5
            tasks, results := createWorkerPool(numWorkers, handler)
            err = copyVersion(printer, log, state.History, src, t, n, handler, tasks, results)
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
    numWorkers := 5
    tasks, results := createWorkerPool(numWorkers, handler)
    return copyVersion(common.AssurePrinter(printer), log, hist, src, t, src.GetDescriptor().Copy(), handler, tasks, results)
}


func copyVersion(printer common.Printer, log logging.Logger, hist common.History, src ocm.ComponentVersionAccess, t ocm.ComponentVersionAccess, prep *compdesc.ComponentDescriptor, handler TransferHandler, tasks chan TransferTask, results chan error) (rerr error) {
    var finalize finalizer.Finalizer

    defer errors.PropagateError(&rerr, finalize.Finalize)

    if handler == nil {
        handler = standard.NewDefaultHandler(nil)
    }

    *t.GetDescriptor() = *prep
    log.Info("  transferring resources and sources")

    // Transfer resources and sources in parallel
    go func() {
        for i, r := range src.GetResources() {
            a, err := r.Access()
            if err != nil {
                results <- errors.Wrapf(err, "%s: accessing resource %d", hist, i)
                continue
            }
            m, err := a.AccessMethod(src)
            if err != nil {
                results <- errors.Wrapf(err, "%s: getting access method for resource %d", hist, i)
                continue
            }
            hint := ocmcpi.ArtifactNameHint(a, src)
            tasks <- TransferTask{Resource: r, Method: m, Hint: hint, Target: t}
        }

        for i, r := range src.GetSources() {
            a, err := r.Access()
            if err != nil {
                results <- errors.Wrapf(err, "%s: accessing source %d", hist, i)
                continue
            }
            m, err := a.AccessMethod(src)
            if err != nil {
                results <- errors.Wrapf(err, "%s: getting access method for source %d", hist, i)
                continue
            }
            hint := ocmcpi.ArtifactNameHint(a, src)
            tasks <- TransferTask{Source: r, Method: m, Hint: hint, Target: t}
        }
        close(tasks)
    }()

    // Wait for all transfers to complete
    for range src.GetResources() {
        if err := <-results; err != nil {
            return err
        }
    }

    for range src.GetSources() {
        if err := <-results; err != nil {
            return err
        }
    }

    return nil
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
