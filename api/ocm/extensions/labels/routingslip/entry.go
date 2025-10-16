package routingslip

import (
	"fmt"
	"strings"

	"github.com/mandelsoft/goutils/errors"
	"github.com/opencontainers/go-digest"

	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/extensions/labels/routingslip/internal"
	"ocm.software/ocm/api/tech/signing"
	"ocm.software/ocm/api/tech/signing/norm/jcs"
	"ocm.software/ocm/api/utils/runtime"
)

func AsGenericEntry(u *runtime.UnstructuredTypedObject) *GenericEntry {
	return internal.AsGenericEntry(u)
}

func ToGenericEntry(e Entry) (*GenericEntry, error) {
	return internal.ToGenericEntry(e)
}

func NewGenericEntryWith(typ string, attrs ...interface{}) (*GenericEntry, error) {
	r := map[string]interface{}{}
	i := 0
	for len(attrs) > i {
		n, ok := attrs[i].(string)
		if !ok {
			return nil, errors.ErrInvalid("key type", fmt.Sprintf("%T", attrs[i]))
		}
		r[n] = attrs[i+1]
		i += 2
	}
	return NewGenericEntry(typ, r)
}

func NewGenericEntry(typ string, data interface{}) (*GenericEntry, error) {
	u, err := runtime.ToUnstructuredTypedObject(data)
	if err != nil {
		return nil, err
	}
	if typ != "" {
		u.SetType(typ)
	}
	return AsGenericEntry(u), nil
}

var excludes = signing.MapExcludes{
	"digest":    nil,
	"signature": nil,
}

type HistoryEntries = []HistoryEntry

type HistoryEntry struct {
	Payload   *GenericEntry         `json:"payload"`
	Timestamp metav1.Timestamp      `json:"timestamp"`
	Parent    *digest.Digest        `json:"parent,omitempty"`
	Links     []Link                `json:"links,omitempty"`
	Digest    digest.Digest         `json:"digest"`
	Signature *metav1.SignatureSpec `json:"signature,omitempty"`
}

func (e *HistoryEntry) Normalize() ([]byte, error) {
	return signing.Normalize(jcs.New(), e, excludes)
}

func (e *HistoryEntry) CalculateDigest() (digest.Digest, error) {
	data, err := e.Normalize()
	if err != nil {
		return "", err
	}
	return digest.SHA256.FromBytes(data), nil
}

type Link struct {
	Name   string        `json:"name"`
	Digest digest.Digest `json:"digest"`
}

func (l Link) Compare(o Link) int {
	r := strings.Compare(l.Name, o.Name)
	if r == 0 {
		r = strings.Compare(l.Digest.String(), o.Digest.String())
	}
	return r
}

func CreateEntry(t runtime.VersionedTypedObject) (Entry, error) {
	return internal.CreateEntry(t)
}
