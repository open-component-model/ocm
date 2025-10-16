package hashattr_test

import (
	"encoding/json"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"ocm.software/ocm/api/config"
	"ocm.software/ocm/api/datacontext"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/extensions/attrs/hashattr"
	"ocm.software/ocm/api/tech/signing/hasher/sha512"
	"ocm.software/ocm/api/utils/runtime"
)

const NAME = "test"

var _ = Describe("attribute", func() {
	var cfgctx config.Context
	var ocmctx ocm.Context

	BeforeEach(func() {
		ocmctx = ocm.New(datacontext.MODE_EXTENDED)
		cfgctx = ocmctx.ConfigContext()
	})

	It("marshal/unmarshal", func() {
		cfg := hashattr.New(sha512.Algorithm)
		data := Must(json.Marshal(cfg))

		r := &hashattr.Config{}
		Expect(json.Unmarshal(data, r)).To(Succeed())
		Expect(r).To(Equal(cfg))
	})

	It("decode", func() {
		attr := &hashattr.Attribute{
			DefaultHasher: sha512.Algorithm,
		}

		r := Must(hashattr.AttributeType{}.Decode([]byte(sha512.Algorithm), runtime.DefaultYAMLEncoding))
		Expect(r).To(Equal(attr))
	})

	It("applies string", func() {
		MustBeSuccessful(cfgctx.GetAttributes().SetAttribute(hashattr.ATTR_KEY, sha512.Algorithm))
		attr := hashattr.Get(ocmctx)
		Expect(attr.GetHasher(ocmctx)).To(Equal(sha512.Handler{}))
	})

	It("applies config", func() {
		cfg := hashattr.New(sha512.Algorithm)

		MustBeSuccessful(cfgctx.ApplyConfig(cfg, "from test"))
		Expect(hashattr.Get(ocmctx).GetHasher(ocmctx)).To(Equal(sha512.Handler{}))
	})
})
