// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package rsakeypair_test

import (
	"bytes"
	"encoding/pem"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/cmds/ocm/testhelper"
	. "github.com/open-component-model/ocm/pkg/testutils"

	"github.com/opencontainers/go-digest"

	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/signingattr"
	"github.com/open-component-model/ocm/pkg/encrypt"
	"github.com/open-component-model/ocm/pkg/signing"
	"github.com/open-component-model/ocm/pkg/signing/handlers/rsa"
)

const ISSUER = "mandelsoft"
const KEYNAME = "test"

var _ = Describe("Test Environment", func() {
	var env *TestEnv
	var defaultContext credentials.Context

	BeforeEach(func() {
		env = NewTestEnv()
		defaultContext = credentials.New()
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("create key pair", func() {

		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("create", "rsakeypair", "key.priv")).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(`
created rsa key pair key.priv[key.pub]
`))
		priv, err := env.ReadFile("key.priv")
		Expect(err).To(Succeed())
		pub, err := env.ReadFile("key.pub")
		Expect(err).To(Succeed())

		d := digest.FromBytes([]byte("digest"))
		sig, err := rsa.Handler{}.Sign(defaultContext, d.Hex(), 0, ISSUER, priv)
		Expect(err).To(Succeed())
		Expect(sig.Algorithm).To(Equal(rsa.Algorithm))
		Expect(sig.MediaType).To(Equal(rsa.MediaType))
		Expect(sig.Issuer).To(Equal(ISSUER))

		err = rsa.Handler{}.Verify(d.Hex(), 0, sig, pub)
		Expect(err).To(Succeed())
	})

	It("create self-signed key pair", func() {

		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("create", "rsakeypair", "key.priv", "CN=mandelsoft")).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(`
created rsa key pair key.priv[key.cert]
`))
		priv, err := env.ReadFile("key.priv")
		Expect(err).To(Succeed())
		pub, err := env.ReadFile("key.cert")
		Expect(err).To(Succeed())

		d := digest.FromBytes([]byte("digest"))
		sig, err := rsa.Handler{}.Sign(defaultContext, d.Hex(), 0, "mandelsoft", priv)
		Expect(err).To(Succeed())
		Expect(sig.Algorithm).To(Equal(rsa.Algorithm))
		Expect(sig.MediaType).To(Equal(rsa.MediaType))
		Expect(sig.Issuer).To(Equal(ISSUER))

		err = rsa.Handler{}.Verify(d.Hex(), 0, sig, pub)
		Expect(err).To(Succeed())
	})

	Context("encryption", func() {
		It("creates encrypted key with new encryption key", func() {
			buf := bytes.NewBuffer(nil)
			Expect(env.CatchOutput(buf).Execute("create", "rsakeypair", "-E", "key.priv")).To(Succeed())
			Expect(buf.String()).To(StringEqualTrimmedWithContext(`
created encrypted rsa key pair key.priv[key.pub][key.priv.ekey]
`))
			pub := Must(env.ReadFile("key.pub"))
			Expect(pub).NotTo(BeNil())

			priv := Must(env.ReadFile("key.priv"))
			Expect(priv).NotTo(BeNil())
			block, rest := pem.Decode(priv)
			Expect(len(rest)).To(Equal(0))
			Expect(block).NotTo(BeNil())
			Expect(block.Type).To(Equal(encrypt.PEM_ENCRYPTED_DATA))

			ekey := Must(env.ReadFile("key.priv.ekey"))
			block, rest = pem.Decode(ekey)
			Expect(len(rest)).To(Equal(0))
			Expect(block).NotTo(BeNil())
			Expect(block.Type).To(Equal(encrypt.PEM_ENCRYPTION_KEY))

			reg := signingattr.Get(env)
			reg.RegisterPrivateKey(KEYNAME, priv)
			reg.RegisterPrivateKey(signing.DecryptionKeyName(KEYNAME), ekey)

			key := Must(signing.ResolvePrivateKey(reg, KEYNAME))
			Expect(key).NotTo(BeNil())

			d := digest.FromBytes([]byte("digest"))
			Must(rsa.Handler{}.Sign(defaultContext, d.Hex(), 0, "mandelsoft", key))

			buf.Reset()
			Expect(env.CatchOutput(buf).Execute("create", "rsakeypair", "-e", KEYNAME, "other.priv")).To(Succeed())
			Expect(buf.String()).To(StringEqualTrimmedWithContext(`
created encrypted rsa key pair other.priv[other.pub]
`))
			pub = Must(env.ReadFile("other.pub"))
			Expect(pub).NotTo(BeNil())

			priv = Must(env.ReadFile("other.priv"))
			Expect(priv).NotTo(BeNil())
			block, rest = pem.Decode(priv)
			Expect(len(rest)).To(Equal(0))
			Expect(block).NotTo(BeNil())
			Expect(block.Type).To(Equal(encrypt.PEM_ENCRYPTED_DATA))
		})
	})
})
