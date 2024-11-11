package internal_test

import (
	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"sigs.k8s.io/yaml"

	"ocm.software/ocm/api/credentials/identity/hostpath"
	"ocm.software/ocm/api/credentials/internal"
)

var _ = Describe("unmarshal cunsomer identity", func() {
	It("with int", func() {
		data := `
scheme: http
hostname: 127.0.0.1
port: 5001
`
		cid := internal.ConsumerIdentity{}
		MustBeSuccessful(yaml.Unmarshal([]byte(data), &cid))
		Expect(cid[hostpath.ID_SCHEME]).To(Equal("http"))
		Expect(cid[hostpath.ID_HOSTNAME]).To(Equal("127.0.0.1"))
		Expect(cid[hostpath.ID_PORT]).To(Equal("5001"))
	})
	It("with float", func() {
		data := `
scheme: http
hostname: 127.0.0.1
port: 3.14
`
		cid := internal.ConsumerIdentity{}
		MustBeSuccessful(yaml.Unmarshal([]byte(data), &cid))
		Expect(cid[hostpath.ID_SCHEME]).To(Equal("http"))
		Expect(cid[hostpath.ID_HOSTNAME]).To(Equal("127.0.0.1"))
		Expect(cid[hostpath.ID_PORT]).To(Equal("3.14"))
	})
	It("with bool", func() {
		data := `
scheme: http
hostname: 127.0.0.1
port: true
`
		cid := internal.ConsumerIdentity{}
		MustBeSuccessful(yaml.Unmarshal([]byte(data), &cid))
		Expect(cid[hostpath.ID_SCHEME]).To(Equal("http"))
		Expect(cid[hostpath.ID_HOSTNAME]).To(Equal("127.0.0.1"))
		Expect(cid[hostpath.ID_PORT]).To(Equal("true"))
	})
	It("with complex value", func() {
		data := `
scheme: http
hostname: 127.0.0.1
port:
  pre: 50
  post: 01
`
		cid := internal.ConsumerIdentity{}
		Expect(yaml.Unmarshal([]byte(data), &cid)).NotTo(Succeed())
	})
	It("with slice value", func() {
		data := `
scheme: http
hostname: 127.0.0.1
port:
- 50
- 01
`
		cid := internal.ConsumerIdentity{}
		Expect(yaml.Unmarshal([]byte(data), &cid)).NotTo(Succeed())
	})
	It("with nil", func() {
		data := `
scheme: http
hostname: 127.0.0.1
port:
`
		id := internal.ConsumerIdentity{
			"scheme":   "http",
			"hostname": "127.0.0.1",
			"port":     "",
		}
		cid := internal.ConsumerIdentity{}
		Expect(yaml.Unmarshal([]byte(data), &cid)).To(Succeed())
		Expect(cid).To(Equal(id))
	})
})
