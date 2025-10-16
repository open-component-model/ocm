package routingslip

import (
	"crypto/x509/pkix"
	"fmt"
	"slices"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/set"
	"github.com/opencontainers/go-digest"
	"ocm.software/ocm/api/ocm/compdesc"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/extensions/attrs/signingattr"
	"ocm.software/ocm/api/tech/signing"
	"ocm.software/ocm/api/tech/signing/hasher/sha256"
	"ocm.software/ocm/api/tech/signing/signutils"
)

const (
	KIND_ENTRY        = "routing slip entry"
	KIND_ENTRY_TYPE   = "routing slip entry type"
	KIND_ROUTING_SLIP = "routing slip"
)

type RoutingSlipIndex map[digest.Digest]*HistoryEntry

func (s RoutingSlipIndex) Leaves() []digest.Digest {
	found := set.Set[digest.Digest]{}
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

func (s RoutingSlipIndex) Verify(ctx Context, name string, issuer *pkix.Name, sig bool, acc SlipAccess) error {
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
			sctx := &signing.DefaultSigningContext{
				Hash:      sha256.Handler{}.Crypto(),
				PublicKey: key,
				Issuer:    issuer,
			}
			err := handler.Verify(last.Digest.Encoded(), last.Signature.ConvertToSigning(), sctx)
			if err != nil {
				return errors.Wrapf(err, "cannot verify entry %s", d)
			}
		}
	}

	found := set.Set[digest.Digest]{}
	for _, id := range leaves {
		s.verify(ctx, name, id, acc, found)
	}
	return nil
}

func (s RoutingSlipIndex) verify(ctx Context, name string, id digest.Digest, acc SlipAccess, found set.Set[digest.Digest]) error {
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
				slip, err := acc.Get(l.Name)
				if err != nil {
					return errors.ErrInvalidWrap(err, KIND_ROUTING_SLIP, l.Name)
				}
				if slip == nil {
					return errors.ErrNotFound(KIND_ROUTING_SLIP, l.Name)
				}
				err = slip.Index().verify(ctx, l.Name, l.Digest, acc, found)
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
	name    string
	issuer  pkix.Name
	entries []HistoryEntry

	index  RoutingSlipIndex
	access SlipAccess
}

func NewRoutingSlip(name string, acc LabelValue) (*RoutingSlip, error) {
	var entries HistoryEntries
	dn, err := signutils.ParseDN(name)
	if err != nil {
		return nil, err
	}
	name = signutils.NormalizeDN(*dn)
	if acc != nil {
		entries = acc[name]
	}
	if err != nil {
		return nil, err
	}
	index := RoutingSlipIndex{}
	for i := range entries {
		index[entries[i].Digest] = &entries[i]
	}
	return &RoutingSlip{name: name, issuer: *dn, access: acc, entries: entries, index: index}, nil
}

func (s *RoutingSlip) GetName() string {
	return s.name
}

func (s *RoutingSlip) Entries() HistoryEntries {
	return slices.Clone(s.entries)
}

func (s *RoutingSlip) Len() int {
	return len(s.entries)
}

func (s *RoutingSlip) Get(i int) *HistoryEntry {
	return &s.entries[i]
}

func (s *RoutingSlip) Leaves() []digest.Digest {
	return s.index.Leaves()
}

func (s *RoutingSlip) Lookup(d digest.Digest) *HistoryEntry {
	return s.index[d]
}

func (s *RoutingSlip) Index() RoutingSlipIndex {
	return s.index
}

func (s *RoutingSlip) Verify(ctx Context, name string, sig bool) error {
	if len(s.entries) == 0 {
		return nil
	}
	return s.index.Verify(ctx, name, &s.issuer, sig, s.access)
}

func (s *RoutingSlip) Add(ctx Context, name string, algo string, e Entry, links []Link, parent ...digest.Digest) (*HistoryEntry, error) {
	registry := signingattr.Get(ctx)
	handler := registry.GetSigner(algo)
	if handler == nil {
		return nil, errors.ErrUnknown(compdesc.KIND_SIGN_ALGORITHM, algo)
	}

	dn, err := signutils.ParseDN(name)
	if err != nil {
		return nil, err
	}
	name = signutils.NormalizeDN(*dn)

	key, err := signing.ResolvePrivateKey(registry, name)
	if err != nil {
		return nil, err
	}
	if key == nil {
		return nil, errors.ErrUnknown(compdesc.KIND_PRIVATE_KEY, name)
	}
	pub := registry.GetPublicKey(name)

	err = s.Verify(ctx, name, true)
	if err != nil {
		return nil, err
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
			last := &s.entries[s.Len()-1]
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
		Digest:    "",
	}

	if base != nil {
		entry.Parent = &base.Digest
		if entry.Parent.String() == "" {
			return nil, fmt.Errorf("no parent digest set")
		}
	}

	for _, l := range links {
		slip, err := s.access.Get(l.Name)
		if err != nil {
			return nil, errors.ErrInvalidWrap(err, KIND_ROUTING_SLIP, l.Name)
		}
		if slip == nil {
			return nil, errors.ErrNotFound(KIND_ROUTING_SLIP, l.Name)
		}
		err = slip.Verify(ctx, name, true)
		if err != nil {
			return nil, err
		}
		if base == nil || (l.Digest != base.Digest && l.Name != s.name) {
			entry.Links = append(entry.Links, l)
		}
	}

	d, err := entry.CalculateDigest()
	if err != nil {
		return nil, err
	}
	entry.Digest = d

	sctx := &signing.DefaultSigningContext{
		Hash:       sha256.Handler{}.Crypto(),
		PrivateKey: key,
		PublicKey:  pub,
		RootCerts:  nil,
		Issuer:     dn,
	}
	sig, err := handler.Sign(ctx.CredentialsContext(), d.Encoded(), sctx)
	if err != nil {
		return nil, err
	}
	if base != nil {
		// keep signatures for leaves, only.
		base.Signature = nil
	}
	entry.Signature = metav1.SignatureSpecFor(sig)
	s.entries = append(s.entries, *entry)
	s.index[entry.Digest] = entry
	return entry, nil
}

////////////////////////////////////////////////////////////////////////////////

func GetSlip(cv cpi.ComponentVersionAccess, name string) (*RoutingSlip, error) {
	label, err := Get(cv)
	if err != nil {
		return nil, err
	}
	return NewRoutingSlip(name, label)
}

func SetSlip(cv cpi.ComponentVersionAccess, slip *RoutingSlip) error {
	label, err := Get(cv)
	if err != nil {
		return err
	}
	if label == nil {
		label = LabelValue{}
	}
	label[slip.name] = slip.entries
	return Set(cv, label)
}
