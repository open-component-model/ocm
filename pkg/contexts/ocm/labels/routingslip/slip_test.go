// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package routingslip_test

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/pkg/testutils"

	"github.com/opencontainers/go-digest"
	"sigs.k8s.io/yaml"

	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/labels/routingslip"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/labels/routingslip/types/comment"
	"github.com/open-component-model/ocm/pkg/env/builder"
	"github.com/open-component-model/ocm/pkg/signing/handlers/rsa"
)

const ORG = "acme.org"
const OTHER = "ocm.software"

var _ = Describe("management", func() {
	var env *builder.Builder

	BeforeEach(func() {
		env = builder.NewBuilder()
		env.RSAKeyPair(ORG, OTHER)
	})

	It("normalizes", func() {
		e := comment.New("start of routing slip")

		parent := digest.Digest("yyy")
		t := metav1.NewTimestampFor(time.Unix(0, 0))

		h := &routingslip.HistoryEntry{
			Payload:   Must(routingslip.ToGenericEntry(e)),
			Timestamp: t,
			Parent:    &parent,
			Digest:    "xxx",
			Signature: metav1.SignatureSpec{},
		}

		fmt.Printf("timestamp: %s\n", t)
		fmt.Printf("(pointer): %s\n", &t)
		Expect(h.Normalize()).To(Equal([]uint8(`{"parent":"yyy","payload":{"comment":"start of routing slip","type":"comment"},"timestamp":"1970-01-01T00:00:00Z"}`)))
	})

	It("adds entry", func() {
		slip := routingslip.NewRoutingSlip(ORG, nil)

		e1 := comment.New("start of routing slip")
		e2 := comment.New("next comment")
		MustBeSuccessful(slip.Add(env.OCMContext(), ORG, rsa.Algorithm, e1, nil))
		MustBeSuccessful(slip.Add(env.OCMContext(), ORG, rsa.Algorithm, e2, nil))

		fmt.Printf("%s\n", string(Must(yaml.Marshal(slip))))

		Expect(slip.Len()).To(Equal(2))
		Expect(slip.Get(1).Parent).To(Equal(&slip.Get(0).Digest))
		MustBeSuccessful(slip.Verify(env.OCMContext(), ORG, true))
	})

	It("adds linked entry", func() {
		label := routingslip.LabelValue{}

		slip := routingslip.NewRoutingSlip(ORG, label)
		label.Set(slip)
		lslip := routingslip.NewRoutingSlip(OTHER, label)
		label.Set(lslip)

		e1 := comment.New("start of routing slip")
		e2 := comment.New("next comment")
		e3 := comment.New("linked comment")

		MustBeSuccessful(slip.Add(env.OCMContext(), ORG, rsa.Algorithm, e1, nil))
		MustBeSuccessful(slip.Add(env.OCMContext(), ORG, rsa.Algorithm, e2, nil))

		d := slip.Get(1).Digest
		MustBeSuccessful(lslip.Add(env.OCMContext(), OTHER, rsa.Algorithm, e3, []routingslip.Link{{Name: ORG, Digest: d}}))

		Expect(lslip.Len()).To(Equal(1))
		Expect(lslip.Get(0).Links).To(Equal([]routingslip.Link{{Name: ORG, Digest: d}}))
	})
})
