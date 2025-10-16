package routingslip_test

import (
	"fmt"
	"time"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/opencontainers/go-digest"
	"ocm.software/ocm/api/helper/builder"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/extensions/labels/routingslip"
	"ocm.software/ocm/api/ocm/extensions/labels/routingslip/types/comment"
	"ocm.software/ocm/api/tech/signing/handlers/rsa"
	"sigs.k8s.io/yaml"
)

const (
	ORG   = "acme.org"
	OTHER = "ocm.software"
)

var _ = Describe("management", func() {
	var env *builder.Builder

	BeforeEach(func() {
		env = builder.NewBuilder()
		env.RSAKeyPair(ORG, OTHER)
	})

	AfterEach(func() {
		env.Cleanup()
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
		}

		fmt.Printf("timestamp: %s\n", t)
		fmt.Printf("(pointer): %s\n", &t)
		Expect(h.Normalize()).To(Equal([]uint8(`{"parent":"yyy","payload":{"comment":"start of routing slip","type":"comment"},"timestamp":"1970-01-01T00:00:00Z"}`)))
	})

	It("adds entry", func() {
		slip := Must(routingslip.NewRoutingSlip(ORG, nil))

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

		slip := Must(routingslip.NewRoutingSlip(ORG, label))
		label.Set(slip)
		lslip := Must(routingslip.NewRoutingSlip(OTHER, label))
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
