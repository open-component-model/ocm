// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package transfer

import (
	"fmt"

	"github.com/mandelsoft/logging"

	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/none"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	ocmcpi "github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/transfer/internal"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/transfer/transferhandler/standard"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/runtime"
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

func transferVersion(printer common.Printer, log logging.Logger, state WalkingState, src ocmcpi.ComponentVersionAccess, tgt ocmcpi.Repository, handler TransferHandler) error {
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

	d := src.GetDescriptor()

	comp, err := tgt.LookupComponent(src.GetName())
	if err != nil {
		return errors.Wrapf(err, "%s: lookup target component", state.History)
	}
	defer comp.Close()

	var ok bool
	t, err := comp.LookupVersion(src.GetVersion())
	defer accessio.Close(t)

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
			defer accessio.Close(t)
		}
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
					// TODO: option to precalculate missing digests (as pre equivalent step).
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
			cv.Close()
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

		// just to be sure: both modes set to false would produce
		// corrupted content in target.
		// If no copy is done, merge must keep the access methods in target!!!
		if !doMerge || doCopy {
			err = copyVersion(printer, log, state.History, src, t, n, handler)
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
	return copyVersion(printer, log, hist, src, t, src.GetDescriptor().Copy(), handler)
}

// copyVersion (purely internal) expects an already prepared target comp desc for t given as prep.
func copyVersion(printer common.Printer, log logging.Logger, hist common.History, src ocm.ComponentVersionAccess, t ocm.ComponentVersionAccess, prep *compdesc.ComponentDescriptor, handler TransferHandler) (rerr error) {
	if handler == nil {
		handler = standard.NewDefaultHandler(nil)
	}

	srccd := src.GetDescriptor()
	cur := *t.GetDescriptor()
	*t.GetDescriptor() = *prep
	log.Info("  transferring resources")
	for i, r := range src.GetResources() {
		var m ocm.AccessMethod

		a, err := r.Access()
		if err == nil {
			m, err = r.AccessMethod()
		}
		if err == nil {
			ok := a.IsLocal(src.GetContext())
			if !ok {
				if !none.IsNone(a.GetKind()) {
					ok, err = handler.TransferResource(src, a, r)
					if err == nil && !ok {
						log.Info("transport omitted", "resource", r.Meta().Name, "index", i, "access", a.GetType())
					}
				}
			}
			if ok {
				var old compdesc.Resource

				hint := ocmcpi.ArtifactNameHint(a, src)
				old, err = cur.GetResourceByIdentity(r.Meta().GetIdentity(srccd.Resources))

				changed := err != nil || old.Digest == nil || !old.Digest.Equal(r.Meta().Digest)
				valueNeeded := err == nil && needsTransport(src.GetContext(), r, &old)
				if changed || valueNeeded {
					var msgs []interface{}
					if !errors.IsErrNotFound(err) {
						if err != nil {
							return err
						}
						if !changed && valueNeeded {
							msgs = []interface{}{"copy"}
						} else {
							msgs = []interface{}{"overwrite"}
						}
					}
					notifyArtifactInfo(printer, log, "resource", i, r.Meta(), hint, msgs...)
					err = handler.HandleTransferResource(r, m, hint, t)
				} else {
					if err == nil { // old resource found -> keep current access method
						t.SetResource(r.Meta(), old.Access, ocm.ModifyResource(), ocm.SkipVerify())
					}
					notifyArtifactInfo(printer, log, "resource", i, r.Meta(), hint, "already present")
				}
			}
			err = errors.Join(err, m.Close())
		}
		if err != nil {
			if !errors.IsErrUnknownKind(err, errors.KIND_ACCESSMETHOD) {
				return errors.Wrapf(err, "%s: transferring resource %d", hist, i)
			}
			printer.Printf("WARN: %s: transferring resource %d: %s (enforce transport by reference)\n", hist, i, err)
		}
	}

	log.Info("  transferring sources")
	for i, r := range src.GetSources() {
		var m ocm.AccessMethod

		a, err := r.Access()
		if err == nil {
			m, err = r.AccessMethod()
		}
		if err == nil {
			ok := a.IsLocal(src.GetContext())
			if !ok {
				if !none.IsNone(a.GetKind()) {
					ok, err = handler.TransferSource(src, a, r)
					if err == nil && !ok {
						log.Info("transport omitted", "source", r.Meta().Name, "index", i, "access", a.GetType())
					}
				}
			}
			if ok {
				// sources do not have digests fo far, so they have to copied, always.
				hint := ocmcpi.ArtifactNameHint(a, src)
				notifyArtifactInfo(printer, log, "source", i, r.Meta(), hint)
				err = errors.Join(err, handler.HandleTransferSource(r, m, hint, t))
			}
			err = errors.Join(err, m.Close())
		}
		if err != nil {
			if !errors.IsErrUnknownKind(err, errors.KIND_ACCESSMETHOD) {
				return errors.Wrapf(err, "%s: transferring source %d", hist, i)
			}
			printer.Printf("WARN: %s: transferring source %d: %s (enforce transport by reference)\n", hist, i, err)
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
