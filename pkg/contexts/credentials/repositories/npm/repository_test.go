package npm_test

import (
	"encoding/json"
	"reflect"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/pkg/testutils"

	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	npmCredentials "github.com/open-component-model/ocm/pkg/contexts/credentials/builtin/npm/identity"
	"github.com/open-component-model/ocm/pkg/contexts/credentials/cpi"
	local "github.com/open-component-model/ocm/pkg/contexts/credentials/repositories/npm"
	"github.com/open-component-model/ocm/pkg/runtimefinalizer"
)

var _ = Describe("NPM config - .npmrc", func() {

	props := common.Properties{
		npmCredentials.ATTR_TOKEN: "npm_TOKEN",
	}

	props2 := common.Properties{
		npmCredentials.ATTR_TOKEN: "bearer_TOKEN",
	}

	var DefaultContext credentials.Context

	BeforeEach(func() {
		DefaultContext = credentials.New()
	})

	specdata := "{\"type\":\"NPMConfig\",\"npmrcFile\":\"testdata/.npmrc\"}"

	It("serializes repo spec", func() {
		spec := local.NewRepositorySpec("testdata/.npmrc")
		data := Must(json.Marshal(spec))
		Expect(data).To(Equal([]byte(specdata)))
	})

	It("deserializes repo spec", func() {
		spec := Must(DefaultContext.RepositorySpecForConfig([]byte(specdata), nil))
		Expect(reflect.TypeOf(spec).String()).To(Equal("*npm.RepositorySpec"))
		Expect(spec.(*local.RepositorySpec).NpmrcFile).To(Equal("testdata/.npmrc"))
	})

	It("resolves repository", func() {
		repo := Must(DefaultContext.RepositoryForConfig([]byte(specdata), nil))
		Expect(reflect.TypeOf(repo).String()).To(Equal("*npm.Repository"))
	})

	It("retrieves credentials", func() {
		repo := Must(DefaultContext.RepositoryForConfig([]byte(specdata), nil))

		creds := Must(repo.LookupCredentials("registry.npmjs.org"))
		Expect(creds.Properties()).To(Equal(props))

		creds = Must(repo.LookupCredentials("npm.registry.acme.com/api/npm"))
		Expect(creds.Properties()).To(Equal(props2))
	})

	It("can access the default context", func() {
		ctx := credentials.New()

		r := runtimefinalizer.GetRuntimeFinalizationRecorder(ctx)
		Expect(r).NotTo(BeNil())

		Must(ctx.RepositoryForConfig([]byte(specdata), nil))

		ci := cpi.NewConsumerIdentity(npmCredentials.CONSUMER_TYPE)
		Expect(ci).NotTo(BeNil())
		credentials := Must(cpi.CredentialsForConsumer(ctx.CredentialsContext(), ci))
		Expect(credentials).NotTo(BeNil())
		Expect(credentials.Properties()).To(Equal(props))
	})
})
