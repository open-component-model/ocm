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
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/labels/routingslip"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/labels/routingslip/entrytypes/comment"
	"github.com/open-component-model/ocm/pkg/env/builder"
	"github.com/open-component-model/ocm/pkg/signing/handlers/rsa"
	. "github.com/open-component-model/ocm/pkg/testutils"
	"github.com/opencontainers/go-digest"
	"sigs.k8s.io/yaml"
)

const ORG = "acme.org"

var _ = Describe("management", func() {
	var env *builder.Builder

	BeforeEach(func() {
		env = builder.NewBuilder()
		env.RSAKeyPair(ORG)
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
		Expect(h.Normalize()).To(Equal([]uint8(`{"parent":"yyy","payload":{"comment":"start of routing slip","type":"comment"},"timestamp":"1970-01-01T01:00:00Z"}`)))
	})

	It("adds entry", func() {
		slip := routingslip.NewRoutingSlip(ORG)

		e1 := comment.New("start of routing slip")
		e2 := comment.New("next comment")
		MustBeSuccessful(slip.Add(env.OCMContext(), ORG, rsa.Algorithm, e1))
		MustBeSuccessful(slip.Add(env.OCMContext(), ORG, rsa.Algorithm, e2))

		fmt.Printf("%s\n", string(Must(yaml.Marshal(slip))))

		Expect(slip.Len()).To(Equal(2))
		Expect(slip.Get(1).Parent).To(Equal(&slip.Get(0).Digest))
		MustBeSuccessful(slip.Verify(env.OCMContext(), ORG, true))
	})
})
