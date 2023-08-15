// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package transfer

import (
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/errors"
)

func PrepareDescriptor(ctx ocm.Context, s *compdesc.ComponentDescriptor, t *compdesc.ComponentDescriptor) (*compdesc.ComponentDescriptor, error) {
	if ctx == nil {
		ctx = ocm.DefaultContext()
	}

	n := s.Copy()
	err := MergeSignatures(t.Signatures, &n.Signatures)
	if err == nil {
		err = MergeLabels(ctx, t.Labels, &n.Labels)
	}
	if err == nil {
		err = MergeLabels(ctx, t.Provider.Labels, &n.Provider.Labels)
	}
	if err == nil {
		err = MergeElements(ctx, t.Sources, n.Sources)
	}
	if err == nil {
		err = MergeElements(ctx, t.Resources, n.Resources)
	}
	if err == nil {
		err = MergeElements(ctx, t.References, n.References)
	}

	if err != nil {
		return nil, err
	}
	return n, nil
}

func MergeElements(ctx ocm.Context, s compdesc.ElementAccessor, t compdesc.ElementAccessor) error {
	for i := 0; i < s.Len(); i++ {
		es := s.Get(i)
		id := es.GetMeta().GetIdentity(s)
		et := compdesc.GetByIdentity(t, id)
		if et != nil {
			if err := MergeLabels(ctx, es.GetMeta().Labels, &et.GetMeta().Labels); err != nil {
				return err
			}

			// keep access for same digest
			if aes, ok := es.(compdesc.ElementArtifactAccessor); ok {
				if des, ok := es.(compdesc.ElementDigestAccessor); ok {
					if des.GetDigest().Equal(et.(compdesc.ElementDigestAccessor).GetDigest()) {
						et.(compdesc.ElementArtifactAccessor).SetAccess(aes.GetAccess())
					}
				}
			}
		}
	}
	return nil
}

// MergeLabels tries to merge old label states into the new target state.
func MergeLabels(ctx ocm.Context, s metav1.Labels, t *metav1.Labels) error {
	for _, l := range s {
		if l.Signing {
			continue
		}
		idx := t.GetIndex(l.Name)
		if idx < 0 {
			*t = append(*t, l)
		} else {
			err := MergeLabel(ctx, l, &(*t)[idx])
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func MergeLabel(ctx ocm.Context, s metav1.Label, t *metav1.Label) error {
	var err error

	n := t.MergeAlgorithm
	if n == "" {
		n = ctx.LabelMergeHandlers().GetAlgorithmFor(s.Name)
	}
	if n != "" {
		h := ctx.LabelMergeHandlers().GetHandler(n)
		if h == nil {
			return errors.ErrUnknown(metav1.KIND_LABEL_MERGE_ALGORITHM, n)
		}
		var cfg ocm.LabelMergeHandlerConfig

		if len(t.MergeConfig) != 0 {
			cfg, err = h.DecodeConfig(t.MergeConfig)
			if err == nil {
				err = cfg.Complete(ctx)
			}
			if err != nil {
				return errors.Wrapf(err, "invalid merge config for label %q", t.Name)
			}
		}
		return h.Merge(ctx, &s, t, cfg)
	}
	return nil
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
