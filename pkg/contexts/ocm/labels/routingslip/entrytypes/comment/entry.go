// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package comment

import (
	"fmt"

	"github.com/open-component-model/ocm/pkg/contexts/ocm/labels/routingslip/spi"
	"github.com/open-component-model/ocm/pkg/runtime"
)

// Type is the access type for a blob in an OCI repository.
const (
	Type   = "comment"
	TypeV1 = Type + runtime.VersionSeparator + "v1"
)

func init() {
	spi.Register(spi.NewEntryType[*Entry](Type, spi.WithDescription(usage)))
	spi.Register(spi.NewEntryType[*Entry](TypeV1))
}

// New creates a new Helm Chart accessor for helm repositories.
func New(comment string) *Entry {
	return &Entry{
		ObjectVersionedType: runtime.NewVersionedTypedObject(Type),
		Comment:             comment,
	}
}

// Entry describes the access for a helm repository.
type Entry struct {
	runtime.ObjectVersionedType `json:",inline"`

	// Comment is just a descriptive text in a routing slip-
	Comment string `json:"comment"`
}

var _ spi.Entry = (*Entry)(nil)

func (a *Entry) Describe(ctx spi.Context) string {
	return fmt.Sprintf("Comment: %s", a.Comment)
}

const usage = `
An unstructured comment as entry in a routing slip.
`
