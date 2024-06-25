package directcreds_test

import (
	"encoding/json"
	"reflect"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/open-component-model/ocm/api/common/common"
	"github.com/open-component-model/ocm/api/credentials"
	"github.com/open-component-model/ocm/api/credentials/extensions/repositories/directcreds"
)

var DefaultContext = credentials.New()

var _ = Describe("direct credentials", func() {
	props := common.Properties{
		"user":     "USER",
		"password": "PASSWORD",
	}
	propsdata := "{\"type\":\"Credentials\",\"properties\":{\"password\":\"PASSWORD\",\"user\":\"USER\"}}"

	It("serializes credentials spec", func() {
		spec := directcreds.NewRepositorySpec(props)
		data, err := json.Marshal(spec)
		Expect(err).To(Succeed())
		Expect(data).To(Equal([]byte(propsdata)))
	})
	It("deserializes credentials spec", func() {
		spec, err := DefaultContext.RepositoryForConfig([]byte(propsdata), nil)
		Expect(err).To(Succeed())
		Expect(reflect.TypeOf(spec).String()).To(Equal("*directcreds.Repository"))
	})

	It("resolved direct credentials", func() {
		spec := directcreds.NewCredentials(props)

		creds, err := DefaultContext.CredentialsForSpec(spec)
		Expect(err).To(Succeed())
		Expect(creds.Properties()).To(Equal(props))
	})
})
