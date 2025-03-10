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

// func transferVersion(printer common.Printer, log logging.Logger, state WalkingState, src ocmcpi.ComponentVersionAccess, tgt ocmcpi.Repository, handler TransferHandler) (rerr error) {
// 	nv := common.VersionedElementKey(src)
// 	log = log.WithValues("history", state.History.String(), "version", nv)
// 	if ok, err := state.Add(ocm.KIND_COMPONENTVERSION, nv); !ok {
// 		return err
// 	}
// 	log.Info("transferring version by US")
// 	printer.Printf("transferring version by US %q...\n", nv)
// 	if handler == nil {
// 		var err error
// 		handler, err = standard.New(standard.Overwrite())
// 		if err != nil {
// 			return err
// 		}
// 	}

// 	var finalize finalizer.Finalizer
// 	defer finalize.FinalizeWithErrorPropagation(&rerr)

// 	d := src.GetDescriptor()

// 	comp, err := tgt.LookupComponent(src.GetName())
// 	if err != nil {
// 		return errors.Wrapf(err, "%s: lookup target component", state.History)
// 	}
// 	finalize.Close(comp, "closing target component")

// 	var ok bool
// 	t, err := comp.LookupVersion(src.GetVersion())
// 	finalize.Close(t, "existing target version")

// 	// references have always to be handled, because of potentially different
// 	// transport modes, which could affect the desired access methods in
// 	// the target environment.

// 	// doTransport controls, whether the transport of the local component
// 	// version has to be re-considered.
// 	doTransport := true

// 	// doMerge controls. whether a potential current version in the target
// 	// environment has to be merged into the transported one.
// 	doMerge := false

// 	// doCopy controls, whether the artifact content has to be considered.
// 	doCopy := true

// 	if err != nil {
// 		if errors.IsErrNotFound(err) {
// 			t, err = comp.NewVersion(src.GetVersion())
// 			finalize.Close(t, "new target version")
// 		}
// 	} else {
// 		ok, err = handler.EnforceTransport(src, t)
// 		if err != nil {
// 			return err
// 		}
// 		if ok {
// 			//  execute transport as if the component version were not present
// 			// 	on the target side.
// 		} else {
// 			// determine transport mode for component version present
// 			// on the target side.
// 			if eq := d.Equivalent(t.GetDescriptor()); eq.IsHashEqual() {
// 				if eq.IsEquivalent() {
// 					if !needsResourceTransport(src, d, t.GetDescriptor(), handler) {
// 						printer.Printf("  version %q already present -> skip transport\n", nv)
// 						doTransport = false
// 					} else {
// 						printer.Printf("  version %q already present -> but requires resource transport\n", nv)
// 					}
// 				} else {
// 					ok, err = handler.UpdateVersion(src, t)
// 					if err != nil {
// 						return err
// 					}
// 					if !ok {
// 						printer.Printf("  version %q requires update of volatile data, but skipped\n", nv)
// 						return nil
// 					}
// 					ok, err = handler.OverwriteVersion(src, t)
// 					if ok {
// 						printer.Printf("  warning: version %q already present, but transport enforced by overwrite option)\n", nv)
// 						doMerge = false
// 						doCopy = true
// 					} else {
// 						printer.Printf("  updating volatile properties of %q\n", nv)
// 						doMerge = true
// 						doCopy = false
// 					}
// 				}
// 			} else {
// 				msg := "  version %q already present, but"
// 				if eq.IsLocalHashEqual() {
// 					if eq.IsArtifactDetectable() {
// 						msg += " differs because some artifact digests are changed"
// 					} else {
// 						// TODO: option to precalculate missing digests (as pre equivalent step).
// 						msg += " might differ, because not all artifact digests are known"
// 					}
// 				} else {
// 					if eq.IsArtifactDetectable() {
// 						if eq.IsArtifactEqual() {
// 							msg += " differs because signature relevant properties have been changed"
// 						} else {
// 							msg += " differs because some artifacts and signature relevant properties have been changed"
// 						}
// 					} else {
// 						msg += "differs because signature relevant properties have been changed (and not all artifact digests are known)"
// 					}
// 				}
// 				ok, err = handler.OverwriteVersion(src, t)
// 				if ok {
// 					doMerge = false
// 					printer.Printf("warning: "+msg+" (transport enforced by overwrite option)\n", nv)
// 				} else {
// 					printer.Printf(msg+" -> transport aborted (use option overwrite option to enforce transport)\n", nv)
// 					return errors.ErrAlreadyExists(ocm.KIND_COMPONENTVERSION, nv.String())
// 				}
// 			}
// 		}
// 	}
// 	if err != nil {
// 		return errors.Wrapf(err, "%s: creating target version", state.History)
// 	}

// 	subp := printer.AddGap("  ")
// 	list := errors.ErrListf("component references for %s", nv)
// 	log.Info("  transferring references")
// 	for _, r := range d.References {
// 		cv, shdlr, err := handler.TransferVersion(src.Repository(), src, &r, tgt)
// 		if err != nil {
// 			return errors.Wrapf(err, "%s: nested component %s[%s:%s]", state.History, r.GetName(), r.ComponentName, r.GetVersion())
// 		}
// 		if cv != nil {
// 			list.Add(transferVersion(subp, log.WithValues("ref", r.Name), state, cv, tgt, shdlr))
// 			list.Addf(nil, cv.Close(), "closing reference %s", r.Name)
// 		}
// 	}

// 	if doTransport {
// 		var n *compdesc.ComponentDescriptor
// 		if doMerge {
// 			log.WithValues("source", src.GetDescriptor(), "target", t.GetDescriptor()).Info("  applying 2-way merge")
// 			n, err = internal.PrepareDescriptor(log, src.GetContext(), src.GetDescriptor(), t.GetDescriptor())
// 			if err != nil {
// 				return err
// 			}
// 		} else {
// 			n = src.GetDescriptor().Copy()
// 		}

// 		var unstr *runtime.UnstructuredTypedObject
// 		if !ocm.IsIntermediate(tgt.GetSpecification()) {
// 			unstr, err = runtime.ToUnstructuredTypedObject(tgt.GetSpecification())
// 			if err != nil {
// 				unstr = nil
// 			}
// 		}
// 		if unstr != nil {
// 			n.RepositoryContexts = append(n.RepositoryContexts, unstr)
// 		}

// 		// just to be sure: both modes set to false would produce
// 		// corrupted content in target.
// 		// If no copy is done, merge must keep the access methods in target!!!
// 		if !doMerge || doCopy {
// 			err = copyVersion(printer, log, state.History, src, t, n, handler)
// 			if err != nil {
// 				return err
// 			}
// 		} else {
// 			*t.GetDescriptor() = *n
// 		}

// 		printer.Printf("...adding component version...\n")
// 		log.Info("  adding component version")
// 		list.Add(comp.AddVersion(t))
// 	}
// 	return list.Result()
// }

// func transferVersion(printer common.Printer, log logging.Logger, state WalkingState, src ocmcpi.ComponentVersionAccess, tgt ocmcpi.Repository, handler TransferHandler) (rerr error) {
//     nv := common.VersionedElementKey(src)
//     log = log.WithValues("history", state.History.String(), "version", nv)

//     if ok, err := state.Add(ocm.KIND_COMPONENTVERSION, nv); !ok {
//         return err
//     }
//     log.Info("transferring version by US")
//     printer.Printf("transferring version by US %q...\n", nv)

//     if handler == nil {
//         var err error
//         handler, err = standard.New(standard.Overwrite())
//         if err != nil {
//             return err
//         }
//     }

//     var finalize finalizer.Finalizer
//     defer finalize.FinalizeWithErrorPropagation(&rerr)

//     d := src.GetDescriptor()

//     comp, err := tgt.LookupComponent(src.GetName())
//     if err != nil {
//         return errors.Wrapf(err, "%s: lookup target component", state.History)
//     }
//     finalize.Close(comp, "closing target component")

//     var ok bool
//     t, err := comp.LookupVersion(src.GetVersion())
//     finalize.Close(t, "existing target version")

//     doTransport := true
//     doMerge := false
//     doCopy := true

//     if err != nil {
//         if errors.IsErrNotFound(err) {
//             t, err = comp.NewVersion(src.GetVersion())
//             finalize.Close(t, "new target version")
//         }
//     } else {
//         ok, err = handler.EnforceTransport(src, t)
//         if err != nil {
//             return err
//         }
//         if ok {
//             // Transport enforced
//         } else {
//             if eq := d.Equivalent(t.GetDescriptor()); eq.IsHashEqual() {
//                 if eq.IsEquivalent() {
//                     if !needsResourceTransport(src, d, t.GetDescriptor(), handler) {
//                         printer.Printf("  version %q already present -> skip transport\n", nv)
//                         doTransport = false
//                     } else {
//                         printer.Printf("  version %q already present -> but requires resource transport\n", nv)
//                     }
//                 } else {
//                     ok, err = handler.UpdateVersion(src, t)
//                     if err != nil {
//                         return err
//                     }
//                     if !ok {
//                         printer.Printf("  version %q requires update of volatile data, but skipped\n", nv)
//                         return nil
//                     }
//                     ok, err = handler.OverwriteVersion(src, t)
//                     if ok {
//                         printer.Printf("  warning: version %q already present, but transport enforced by overwrite option)\n", nv)
//                         doMerge = false
//                         doCopy = true
//                     } else {
//                         printer.Printf("  updating volatile properties of %q\n", nv)
//                         doMerge = true
//                         doCopy = false
//                     }
//                 }
//             } else {
//                 msg := "  version %q already present, but"
//                 if eq.IsLocalHashEqual() {
//                     if eq.IsArtifactDetectable() {
//                         msg += " differs because some artifact digests are changed"
//                     } else {
//                         msg += " might differ, because not all artifact digests are known"
//                     }
//                 } else {
//                     if eq.IsArtifactDetectable() {
//                         if eq.IsArtifactEqual() {
//                             msg += " differs because signature relevant properties have been changed"
//                         } else {
//                             msg += " differs because some artifacts and signature relevant properties have been changed"
//                         }
//                     } else {
//                         msg += "differs because signature relevant properties have been changed (and not all artifact digests are known)"
//                     }
//                 }
//                 ok, err = handler.OverwriteVersion(src, t)
//                 if ok {
//                     doMerge = false
//                     printer.Printf("warning: "+msg+" (transport enforced by overwrite option)\n", nv)
//                 } else {
//                     printer.Printf(msg+" -> transport aborted (use option overwrite option to enforce transport)\n", nv)
//                     return errors.ErrAlreadyExists(ocm.KIND_COMPONENTVERSION, nv.String())
//                 }
//             }
//         }
//     }
//     if err != nil {
//         return errors.Wrapf(err, "%s: creating target version", state.History)
//     }

//     subp := printer.AddGap("  ")
//     list := errors.ErrListf("component references for %s", nv)
//     log.Info("  transferring references")

//     for _, r := range d.References {
//         cv, shdlr, err := handler.TransferVersion(src.Repository(), src, &r, tgt)
//         if err != nil {
//             return errors.Wrapf(err, "%s: nested component %s[%s:%s]", state.History, r.GetName(), r.ComponentName, r.GetVersion())
//         }
//         if cv != nil {
//             list.Add(transferVersion(subp, log.WithValues("ref", r.Name), state, cv, tgt, shdlr))
//             list.Addf(nil, cv.Close(), "closing reference %s", r.Name)
//         }
//     }

//     if doTransport {
//         var n *compdesc.ComponentDescriptor
//         if doMerge {
//             log.WithValues("source", src.GetDescriptor(), "target", t.GetDescriptor()).Info("  applying 2-way merge")
//             n, err = internal.PrepareDescriptor(log, src.GetContext(), src.GetDescriptor(), t.GetDescriptor())
//             if err != nil {
//                 return err
//             }
//         } else {
//             n = src.GetDescriptor().Copy()
//         }

//         var unstr *runtime.UnstructuredTypedObject
//         if !ocm.IsIntermediate(tgt.GetSpecification()) {
//             unstr, err = runtime.ToUnstructuredTypedObject(tgt.GetSpecification())
//             if err != nil {
//                 unstr = nil
//             }
//         }
//         if unstr != nil {
//             n.RepositoryContexts = append(n.RepositoryContexts, unstr)
//         }

//         if !doMerge || doCopy {
//             numWorkers := 5
//             tasks, results := createWorkerPool(numWorkers, handler)
//             err = copyVersion(printer, log, state.History, src, t, n, handler, tasks, results)
//             if err != nil {
//                 return err
//             }
//         } else {
//             *t.GetDescriptor() = *n
//         }

//         printer.Printf("...adding component version...\n")
//         log.Info("  adding component version")
//         list.Add(comp.AddVersion(t))
//     }
//     return list.Result()
// }

type TransferTask struct {
    Resource ocm.ResourceAccess
    Source   ocm.SourceAccess
    Method   ocmcpi.AccessMethod
    Hint     string
    Target   ocm.ComponentVersionAccess
}

// type ResourceTransferTask struct {
//     Resource ocm.ResourceAccess
//     Method   ocmcpi.AccessMethod
//     Hint     string
//     Target   ocm.ComponentVersionAccess
// }

// type SourceTransferTask struct {
//     Source ocm.SourceAccess
//     Method ocmcpi.AccessMethod
//     Hint   string
//     Target ocm.ComponentVersionAccess
// }

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

// func CopyVersion(printer common.Printer, log logging.Logger, hist common.History, src ocm.ComponentVersionAccess, t ocm.ComponentVersionAccess, handler TransferHandler) (rerr error) {
// 	return copyVersion(common.AssurePrinter(printer), log, hist, src, t, src.GetDescriptor().Copy(), handler)
// }

func CopyVersion(printer common.Printer, log logging.Logger, hist common.History, src ocm.ComponentVersionAccess, t ocm.ComponentVersionAccess, handler TransferHandler) (rerr error) {
    numWorkers := 5
    tasks, results := createWorkerPool(numWorkers, handler)
    return copyVersion(common.AssurePrinter(printer), log, hist, src, t, src.GetDescriptor().Copy(), handler, tasks, results)
}

//copyVersion (purely internal) expects an already prepared target comp desc for t given as prep.
// func copyVersion(printer common.Printer, log logging.Logger, hist common.History, src ocm.ComponentVersionAccess, t ocm.ComponentVersionAccess, prep *compdesc.ComponentDescriptor, handler TransferHandler) (rerr error) {
// 	var finalize finalizer.Finalizer

// 	defer errors.PropagateError(&rerr, finalize.Finalize)

// 	if handler == nil {
// 		handler = standard.NewDefaultHandler(nil)
// 	}

// 	srccd := src.GetDescriptor()
// 	cur := *t.GetDescriptor()
// 	*t.GetDescriptor() = *prep
// 	log.Info("  transferring resources")
// 	for i, r := range src.GetResources() {
// 		var m ocmcpi.AccessMethod

// 		nested := finalize.Nested()

// 		a, err := r.Access()
// 		if err == nil {
// 			m, err = a.AccessMethod(src)
// 			nested.Close(m, fmt.Sprintf("%s: transferring resource %d: closing access method", hist, i))
// 		}
// 		if err == nil {
// 			ok := a.IsLocal(src.GetContext())
// 			if !ok {
// 				if !none.IsNone(a.GetKind()) {
// 					ok, err = handler.TransferResource(src, a, r)
// 					if err == nil && !ok {
// 						log.Info("transport omitted", "resource", r.Meta().Name, "index", i, "access", a.GetType())
// 					}
// 				}
// 			}
// 			if ok {
// 				var old compdesc.Resource

// 				hint := ocmcpi.ArtifactNameHint(a, src)
// 				old, err = cur.GetResourceByIdentity(r.Meta().GetIdentity(srccd.Resources))

// 				changed := err != nil || old.Digest == nil || !old.Digest.Equal(r.Meta().Digest)
// 				valueNeeded := err == nil && needsTransport(src.GetContext(), r, &old)
// 				if changed || valueNeeded {
// 					var msgs []interface{}
// 					if !errors.IsErrNotFound(err) {
// 						if err != nil {
// 							return err
// 						}
// 						if !changed && valueNeeded {
// 							msgs = []interface{}{"copy"}
// 						} else {
// 							msgs = []interface{}{"overwrite"}
// 						}
// 					}
// 					notifyArtifactInfo(printer, log, "resource", i, r.Meta(), hint, msgs...)
// 					err = handler.HandleTransferResource(r, m, hint, t)
// 				} else {
// 					if err == nil { // old resource found -> keep current access method
// 						t.SetResource(r.Meta(), old.Access, ocm.ModifyElement(), ocm.SkipVerify(), ocm.DisableExtraIdentityDefaulting())
// 					}
// 					notifyArtifactInfo(printer, log, "resource", i, r.Meta(), hint, "already present")
// 				}
// 			}
// 		}
// 		if err != nil {
// 			if !errors.IsErrUnknownKind(err, errkind.KIND_ACCESSMETHOD) {
// 				return errors.Wrapf(err, "%s: transferring resource %d", hist, i)
// 			}
// 			printer.Printf("WARN: %s: transferring resource %d: %s (enforce transport by reference)\n", hist, i, err)
// 		}
// 		err = nested.Finalize()
// 		if err != nil {
// 			return err
// 		}
// 	}

// 	log.Info("  transferring sources")
// 	for i, r := range src.GetSources() {
// 		var m ocmcpi.AccessMethod

// 		a, err := r.Access()
// 		if err == nil {
// 			m, err = a.AccessMethod(src)
// 		}
// 		if err == nil {
// 			ok := a.IsLocal(src.GetContext())
// 			if !ok {
// 				if !none.IsNone(a.GetKind()) {
// 					ok, err = handler.TransferSource(src, a, r)
// 					if err == nil && !ok {
// 						log.Info("transport omitted", "source", r.Meta().Name, "index", i, "access", a.GetType())
// 					}
// 				}
// 			}
// 			if ok {
// 				// sources do not have digests so far, so they have to copied, always.
// 				hint := ocmcpi.ArtifactNameHint(a, src)
// 				notifyArtifactInfo(printer, log, "source", i, r.Meta(), hint)
// 				err = errors.Join(err, handler.HandleTransferSource(r, m, hint, t))
// 			}
// 			err = errors.Join(err, m.Close())
// 		}
// 		if err != nil {
// 			if !errors.IsErrUnknownKind(err, errkind.KIND_ACCESSMETHOD) {
// 				return errors.Wrapf(err, "%s: transferring source %d", hist, i)
// 			}
// 			printer.Printf("WARN: %s: transferring source %d: %s (enforce transport by reference)\n", hist, i, err)
// 		}
// 	}
// 	return nil
// }

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
