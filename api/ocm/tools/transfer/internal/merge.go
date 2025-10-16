package internal

import (
	"github.com/mandelsoft/logging"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/compdesc"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/valuemergehandler"
	"ocm.software/ocm/api/ocm/valuemergehandler/hpi"
	"ocm.software/ocm/api/utils/runtime"
)

// PrepareDescriptor provides a descriptor for the transport target based on a
// descriptor from the transport source and a descriptor already prsent at the
// target.
func PrepareDescriptor(log logging.Logger, ctx ocm.Context, s *compdesc.ComponentDescriptor, t *compdesc.ComponentDescriptor) (*compdesc.ComponentDescriptor, error) {
	if ctx == nil {
		ctx = ocm.DefaultContext()
	}

	n := s.Copy()
	err := MergeSignatures(t.Signatures, &n.Signatures)
	if err == nil {
		err = MergeLabels(log, ctx, t.Labels, &n.Labels)
	}
	if err == nil {
		err = MergeLabels(log, ctx, t.Provider.Labels, &n.Provider.Labels)
	}
	if err == nil {
		err = MergeElements(log, ctx, t.Sources, n.Sources)
	}
	if err == nil {
		err = MergeElements(log, ctx, t.Resources, n.Resources)
	}
	if err == nil {
		err = MergeElements(log, ctx, t.References, n.References)
	}

	if err != nil {
		return nil, err
	}
	return n, nil
}

func MergeElements(log logging.Logger, ctx ocm.Context, s compdesc.ElementListAccessor, t compdesc.ElementListAccessor) error {
	for i := 0; i < s.Len(); i++ {
		es := s.Get(i)
		id := es.GetMeta().GetIdentity(s)
		et := compdesc.GetByIdentity(t, id)
		if et != nil {
			labels := et.GetMeta().GetLabels()
			if err := MergeLabels(log, ctx, es.GetMeta().GetLabels(), &labels); err != nil {
				return err
			}
			et.GetMeta().SetLabels(labels)

			// keep access for same digest
			if aes, ok := es.(compdesc.ElementArtifactAccessor); ok {
				if des, ok := es.(compdesc.ElementDigestAccessor); ok {
					if des.GetDigest() != nil && des.GetDigest().Equal(et.(compdesc.ElementDigestAccessor).GetDigest()) {
						et.(compdesc.ElementArtifactAccessor).SetAccess(aes.GetAccess())
					}
				}
			}
			// keep digest for locally signed/hashed elements
			if des, ok := es.(compdesc.ElementDigestAccessor); ok {
				if des.GetDigest() != nil {
					det := et.(compdesc.ElementDigestAccessor)
					if det.GetDigest() == nil {
						det.SetDigest(des.GetDigest())
					}
				}
			}
		}
	}
	return nil
}

// MergeLabels tries to merge old label states into the new target state.
func MergeLabels(log logging.Logger, ctx ocm.Context, s metav1.Labels, t *metav1.Labels) error {
	for _, l := range s {
		if l.Signing {
			continue
		}
		idx := t.GetIndex(l.Name)
		if idx < 0 {
			log.Trace("appending label", "name", l.Name, "value", l.Value)
			*t = append(*t, l)
		} else {
			err := MergeLabel(ctx, l, &(*t)[idx])
			if err != nil {
				return err
			}
			log.Trace("merge result", "name", l.Name, "result", (*t)[idx].Value)
		}
	}
	return nil
}

func MergeLabel(ctx ocm.Context, s metav1.Label, t *metav1.Label) error {
	r := valuemergehandler.Value{t.Value}
	v := t.Version
	if v == "" {
		v = "v1"
	}
	mod, err := valuemergehandler.Merge(ctx, t.Merge, hpi.LabelHint(t.Name, v), runtime.RawValue{s.Value}, &r)
	if mod {
		t.Value = r.RawMessage
	}
	return err
}

// MergeSignatures tries to merge old signatures into the new target state.
func MergeSignatures(s metav1.Signatures, t *metav1.Signatures) error {
	for _, sig := range s {
		idx := t.GetIndex(sig.Name)
		if idx < 0 {
			*t = append(*t, sig)
		}
	}
	return nil
}
