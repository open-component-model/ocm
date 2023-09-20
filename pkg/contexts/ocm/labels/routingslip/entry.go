// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package routingslip

import (
	"fmt"

	"github.com/opencontainers/go-digest"

	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/labels/routingslip/internal"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/runtime"
	"github.com/open-component-model/ocm/pkg/signing"
	"github.com/open-component-model/ocm/pkg/signing/norm/jcs"
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
	Payload   *GenericEntry        `json:"payload"`
	Timestamp metav1.Timestamp     `json:"timestamp"`
	Parent    *digest.Digest       `json:"parent,omitempty"`
	Links     []Link               `json:"links,omitempty"`
	Digest    digest.Digest        `json:"digest"`
	Signature metav1.SignatureSpec `json:"signature"`
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

func CreateEntry(t runtime.VersionedTypedObject) (Entry, error) {
	return internal.CreateEntry(t)
}
