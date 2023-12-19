// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package routingslip_test

import (
	"encoding/json"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/pkg/testutils"

	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/labels/routingslip"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/labels/routingslip/internal"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/labels/routingslip/types/comment"
	"github.com/open-component-model/ocm/pkg/runtime"
)

const TYPE = "my"

type My struct {
	runtime.ObjectVersionedTypedObject
	Value string `json:"value"`
}

func (m *My) Describe(ctx ocm.Context) string {
	return fmt.Sprintf("%s with %q", m.GetKind(), m.Value)
}

func (m *My) Validate(ctx ocm.Context) error {
	return nil
}

func New(v string) *My {
	return &My{
		ObjectVersionedTypedObject: runtime.NewVersionedTypedObject(TYPE),
		Value:                      v,
	}
}

var _ = Describe("routing slip", func() {
	now := metav1.NewTimestamp()

	It("parses", func() {
		e := &routingslip.HistoryEntry{
			Payload:   Must(internal.ToGenericEntry(New("test"))),
			Digest:    "sha:digest",
			Timestamp: now,
			Signature: &metav1.SignatureSpec{
				Algorithm: "algo",
				Value:     "value",
				MediaType: "mime",
				Issuer:    "acme.org",
			},
		}

		data := Must(json.Marshal(e))
		fmt.Printf("%s\n", string(data))

		var r routingslip.HistoryEntry
		MustBeSuccessful(json.Unmarshal(data, &r))

		Expect(&r).To(DeepEqual(e))
	})

	It("parses predefined", func() {
		entry := `{"payload":{"type":"comment","comment":"some comment"},"digest":"sha:digest","timestamp":"2023-08-25T10:39:20+02:00","signature":{"algorithm":"algo","value":"value","mediaType":"mime","issuer":"acme.org"}}`

		var r routingslip.HistoryEntry
		MustBeSuccessful(json.Unmarshal([]byte(entry), &r))

		Expect(Must(r.Payload.Evaluate(ocm.DefaultContext())).(*comment.Entry).Comment).To(Equal("some comment"))
	})
})
