// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package routingslip

import (
	"github.com/opencontainers/go-digest"

	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/labels/routingslip/internal"
	"github.com/open-component-model/ocm/pkg/runtime"
	"github.com/open-component-model/ocm/pkg/signing"
	"github.com/open-component-model/ocm/pkg/signing/norm/jcs"
)

type (
	Context      = internal.Context
	Entry        = internal.Entry
	GenericEntry = internal.GenericEntry
)

func ToGenericEntry(e Entry) (*GenericEntry, error) {
	return internal.ToGenericEntry(e)
}

var excludes = signing.MapExcludes{
	"digest":    nil,
	"signature": nil,
}

type HistoryEntry struct {
	Payload   *GenericEntry        `json:"payload"`
	Timestamp metav1.Timestamp     `json:"timestamp"`
	Parent    *digest.Digest       `json:"parent,omitempty"`
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

func CreateEntry(t runtime.VersionedTypedObject) (Entry, error) {
	return internal.CreateEntry(t)
}
