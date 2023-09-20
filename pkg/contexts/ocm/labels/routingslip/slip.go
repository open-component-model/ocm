// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package routingslip

import (
	"fmt"

	"github.com/opencontainers/go-digest"

	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/signingattr"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/generics"
	"github.com/open-component-model/ocm/pkg/signing"
	"github.com/open-component-model/ocm/pkg/signing/hasher/sha256"
)

const (
	KIND_ENTRY        = "routing slip entry"
	KIND_ENTRY_TYPE   = "routing slip entry type"
	KIND_ROUTING_SLIP = "routing slip"
)

type RoutingSlipIndex map[digest.Digest]*HistoryEntry

func (s RoutingSlipIndex) Leaves() []digest.Digest {
	found := generics.Set[digest.Digest]{}
	for _, e := range s {
		found.Add(e.Digest)
	}
	for _, e := range s {
		if e.Parent != nil {
			found.Delete(*e.Parent)
		}
	}
	return found.AsArray()
}

func (s RoutingSlipIndex) Verify(ctx Context, name string, sig bool, acc SlipAccess) error {
	if len(s) == 0 {
		return nil
	}
	leaves := s.Leaves()

	if sig {
		registry := signingattr.Get(ctx)
		key := registry.GetPublicKey(name)
		if key == nil {
			var err error
			key, err = signing.ResolvePrivateKey(registry, name)
			if err != nil {
				return err
			}
		}
		if key == nil {
			return errors.ErrNotFound(compdesc.KIND_PUBLIC_KEY, name)
		}
		for _, d := range leaves {
			last := s[d]
			handler := registry.GetVerifier(last.Signature.Algorithm)
			if handler == nil {
				return errors.ErrUnknown(compdesc.KIND_VERIFY_ALGORITHM, last.Signature.Algorithm)
			}
			err := handler.Verify(last.Digest.Encoded(), sha256.Handler{}.Crypto(), last.Signature.ConvertToSigning(), key)
			if err != nil {
				return errors.Wrapf(err, "cannot verify entry %s", d)
			}
		}
	}

	found := generics.Set[digest.Digest]{}
	for _, id := range leaves {
		s.verify(ctx, name, id, acc, found)
	}
	return nil
}

func (s RoutingSlipIndex) verify(ctx Context, name string, id digest.Digest, acc SlipAccess, found generics.Set[digest.Digest]) error {
	cur := s[id]
	if cur == nil {
		return errors.ErrNotFound(KIND_ENTRY, id.String(), name)
	}
	for {
		if found.Contains(cur.Digest) {
			return nil
		}
		found.Add(cur.Digest)
		d, err := cur.CalculateDigest()
		if err != nil {
			return err
		}
		if d != cur.Digest {
			return fmt.Errorf("content digest %q does not match %q in %s", d, cur.Digest, name)
		}
		for _, l := range cur.Links {
			if l.Name == name {
				err := s.verify(ctx, name, l.Digest, acc, found)
				if err != nil {
					return err
				}
			} else {
				slip := acc.Get(l.Name)
				if slip == nil {
					return errors.ErrNotFound(KIND_ROUTING_SLIP, l.Name)
				}
				err := slip.Index().verify(ctx, l.Name, l.Digest, acc, found)
				if err != nil {
					return err
				}
			}
		}
		if cur.Parent == nil {
			break
		}
		if cur = s[*cur.Parent]; cur == nil {
			return fmt.Errorf("parent %q of %q not found in %s", cur.Parent, d, name)
		}
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////

type RoutingSlip struct {
	Name    string
	Entries []HistoryEntry
	Access  SlipAccess
}

func NewRoutingSlip(name string, acc SlipAccess) *RoutingSlip {
	return &RoutingSlip{Name: name, Access: acc}
}

func (s *RoutingSlip) Len() int {
	return len(s.Entries)
}

func (s *RoutingSlip) Get(i int) *HistoryEntry {
	return &s.Entries[i]
}

func (s *RoutingSlip) Leaves() []digest.Digest {
	found := generics.Set[digest.Digest]{}
	for _, e := range s.Entries {
		found.Add(e.Digest)
	}
	for _, e := range s.Entries {
		if e.Parent != nil {
			found.Delete(*e.Parent)
		}
	}
	return found.AsArray()
}

func (s *RoutingSlip) Lookup(d digest.Digest) *HistoryEntry {
	for i := range s.Entries {
		if s.Entries[i].Digest == d {
			return &s.Entries[i]
		}
	}
	return nil
}

func (s *RoutingSlip) Index() RoutingSlipIndex {
	index := RoutingSlipIndex{}
	for i := range s.Entries {
		index[s.Entries[i].Digest] = &s.Entries[i]
	}
	return index
}

func (s *RoutingSlip) Verify(ctx Context, name string, sig bool) error {
	if len(s.Entries) == 0 {
		return nil
	}
	return s.Index().Verify(ctx, name, sig, s.Access)
}

func (s *RoutingSlip) Add(ctx Context, name string, algo string, e Entry, links []Link, parent ...digest.Digest) (*HistoryEntry, error) {
	registry := signingattr.Get(ctx)
	handler := registry.GetSigner(algo)
	if handler == nil {
		return nil, errors.ErrUnknown(compdesc.KIND_SIGN_ALGORITHM, algo)
	}
	key, err := signing.ResolvePrivateKey(registry, name)
	if err != nil {
		return nil, err
	}
	if key == nil {
		return nil, errors.ErrUnknown(compdesc.KIND_PRIVATE_KEY, name)
	}

	err = s.Verify(ctx, name, true)
	if err != nil {
		return nil, err
	}

	for _, l := range links {
		slip := s.Access.Get(l.Name)
		if slip == nil {
			return nil, errors.ErrNotFound(KIND_ROUTING_SLIP, l.Name)
		}
		err = slip.Verify(ctx, name, true)
		if err != nil {
			return nil, err
		}
	}

	var base *HistoryEntry
	if len(parent) > 0 {
		base = s.Lookup(parent[0])
		if base == nil {
			return nil, errors.ErrNotFound(KIND_ENTRY, parent[0].String(), name)
		}
	}
	if base == nil && s.Len() > 0 {
		leaves := s.Leaves()
		if len(leaves) == 1 {
			base = s.Lookup(leaves[0])
		} else {
			last := &s.Entries[s.Len()-1]
			for _, l := range leaves {
				if last.Digest == l {
					base = last
					break
				}
			}
			if base == nil {
				return nil, fmt.Errorf("no unique base entry found in %s", name)
			}
		}
	}
	gen, err := ToGenericEntry(e)
	if err != nil {
		return nil, err
	}
	entry := &HistoryEntry{
		Payload:   gen,
		Timestamp: metav1.NewTimestamp(),
		Links:     links,
		Digest:    "",
		Signature: metav1.SignatureSpec{},
	}
	if base != nil {
		entry.Parent = &base.Digest
		if entry.Parent.String() == "" {
			return nil, fmt.Errorf("no parent digest set")
		}
	}
	d, err := entry.CalculateDigest()
	if err != nil {
		return nil, err
	}
	entry.Digest = d

	sig, err := handler.Sign(ctx.CredentialsContext(), d.Encoded(), sha256.Handler{}.Crypto(), name, key)
	if err != nil {
		return nil, err
	}
	entry.Signature = *metav1.SignatureSpecFor(sig)
	s.Entries = append(s.Entries, *entry)
	return entry, nil
}

////////////////////////////////////////////////////////////////////////////////

func GetSlip(cv cpi.ComponentVersionAccess, name string) (*RoutingSlip, error) {
	label, err := Get(cv)
	if err != nil {
		return nil, err
	}
	return &RoutingSlip{
		Name:    name,
		Entries: label[name],
		Access:  label,
	}, nil
}

func SetSlip(cv cpi.ComponentVersionAccess, slip *RoutingSlip) error {
	label, err := Get(cv)
	if err != nil {
		return err
	}
	if label == nil {
		label = LabelValue{}
	}
	label[slip.Name] = slip.Entries
	return Set(cv, label)
}
