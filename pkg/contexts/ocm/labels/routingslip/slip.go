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
	"github.com/open-component-model/ocm/pkg/signing/hasher/sha256"
)

type RoutingSlip []HistoryEntry

func (s RoutingSlip) Leaves() []digest.Digest {
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

func (s RoutingSlip) Lookup(d digest.Digest) *HistoryEntry {
	for i := range s {
		if s[i].Digest == d {
			return &s[i]
		}
	}
	return nil
}

func (s RoutingSlip) Verify(ctx Context, name string, sig bool) error {
	if len(s) == 0 {
		return nil
	}
	leaves := s.Leaves()

	if sig {
		registry := signingattr.Get(ctx)
		key := registry.GetPublicKey(name)
		if key == nil {
			key = registry.GetPrivateKey(name)
		}
		if key == nil {
			return errors.ErrNotFound(compdesc.KIND_PUBLIC_KEY, name)
		}
		for _, d := range leaves {
			last := s.Lookup(d)
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
leaves:
	for _, d := range leaves {
		cur := s.Lookup(d)

		for {
			if found.Contains(cur.Digest) {
				continue leaves
			}
			found.Add(cur.Digest)
			d, err := cur.CalculateDigest()
			if err != nil {
				return err
			}
			if d != cur.Digest {
				return fmt.Errorf("content digest %q dow not match %q", d, cur.Digest)
			}
			if cur.Parent == nil {
				break
			}
			if cur = s.Lookup(*cur.Parent); cur == nil {
				return fmt.Errorf("parent %q of %q not found", cur.Parent, d)
			}
		}
	}
	return nil
}

func (s *RoutingSlip) Add(ctx Context, name string, algo string, e Entry) (*HistoryEntry, error) {
	registry := signingattr.Get(ctx)
	handler := registry.GetSigner(algo)
	if handler == nil {
		return nil, errors.ErrUnknown(compdesc.KIND_SIGN_ALGORITHM, algo)
	}
	key := registry.GetPrivateKey(name)
	if key == nil {
		return nil, errors.ErrUnknown(compdesc.KIND_PRIVATE_KEY, name)
	}

	gen, err := ToGenericEntry(e)
	if err != nil {
		return nil, err
	}
	entry := &HistoryEntry{
		Payload:   gen,
		Timestamp: metav1.NewTimestamp(),
		Digest:    "",
		Signature: metav1.SignatureSpec{},
	}
	if len(*s) > 0 {
		entry.Parent = &(*s)[len(*s)-1].Digest
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
	*s = append(*s, *entry)
	return entry, nil
}

func GetSlip(cv cpi.ComponentVersionAccess, name string) (RoutingSlip, error) {
	label, err := Get(cv)
	if err != nil {
		return nil, err
	}
	return label[name], nil
}

func SetSlip(cv cpi.ComponentVersionAccess, name string, slip RoutingSlip) error {
	label, err := Get(cv)
	if err != nil {
		return err
	}
	if label == nil {
		label = Label{}
	}
	label[name] = slip
	return Set(cv, label)
}
