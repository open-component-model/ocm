// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package dockerconfig_test

import (
	"encoding/json"
	"os"
	"reflect"
	"runtime"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/v2/pkg/testutils"

	"github.com/open-component-model/ocm/v2/pkg/common"
	"github.com/open-component-model/ocm/v2/pkg/contexts/credentials"
	"github.com/open-component-model/ocm/v2/pkg/contexts/credentials/cpi"
	local "github.com/open-component-model/ocm/v2/pkg/contexts/credentials/repositories/dockerconfig"
	"github.com/open-component-model/ocm/v2/pkg/contexts/oci/identity"
	"github.com/open-component-model/ocm/v2/pkg/finalizer"
)

var _ = Describe("docker config", func() {

	props := common.Properties{
		"username":      "mandelsoft",
		"password":      "password",
		"serverAddress": "https://index.docker.io/v1/",
	}

	props2 := common.Properties{
		"username":      "mandelsoft",
		"password":      "token",
		"serverAddress": "https://ghcr.io",
	}

	var DefaultContext credentials.Context

	BeforeEach(func() {
		DefaultContext = credentials.New()
	})

	Context("file based", func() {
		specdata := "{\"type\":\"DockerConfig\",\"dockerConfigFile\":\"testdata/dockerconfig.json\"}"
		specdata2 := "{\"type\":\"DockerConfig\",\"dockerConfigFile\":\"testdata/dockerconfig.json\",\"propagateConsumerIdentity\":true}"

		It("serializes repo spec", func() {
			spec := local.NewRepositorySpec("testdata/dockerconfig.json")
			data := Must(json.Marshal(spec))
			Expect(data).To(Equal([]byte(specdata)))

			spec = local.NewRepositorySpec("testdata/dockerconfig.json").WithConsumerPropagation(true)
			data = Must(json.Marshal(spec))
			Expect(data).To(Equal([]byte(specdata2)))
		})

		It("deserializes repo spec", func() {
			spec := Must(DefaultContext.RepositorySpecForConfig([]byte(specdata), nil))
			Expect(reflect.TypeOf(spec).String()).To(Equal("*dockerconfig.RepositorySpec"))
			Expect(spec.(*local.RepositorySpec).DockerConfigFile).To(Equal("testdata/dockerconfig.json"))
		})

		It("resolves repository", func() {
			repo := Must(DefaultContext.RepositoryForConfig([]byte(specdata), nil))
			Expect(reflect.TypeOf(repo).String()).To(Equal("*dockerconfig.Repository"))
		})

		It("retrieves credentials", func() {
			repo := Must(DefaultContext.RepositoryForConfig([]byte(specdata), nil))

			creds := Must(repo.LookupCredentials("index.docker.io"))
			Expect(creds.Properties()).To(Equal(props))

			creds = Must(repo.LookupCredentials("ghcr.io"))
			Expect(creds.Properties()).To(Equal(props2))
		})

		It("propagates credentials to consumer identity", func() {
			Must(DefaultContext.RepositoryForConfig([]byte(specdata2), nil))

			creds := Must(credentials.CredentialsForConsumer(DefaultContext, credentials.ConsumerIdentity{
				cpi.ATTR_TYPE:        identity.CONSUMER_TYPE,
				identity.ID_HOSTNAME: "ghcr.io",
			}))
			Expect(creds.Properties()).To(Equal(props2))
		})
	})

	Context("inline data", func() {
		specdata := "{\"type\":\"DockerConfig\",\"dockerConfig\":{\"auths\":{\"https://index.docker.io/v1/\":{\"auth\":\"bWFuZGVsc29mdDpwYXNzd29yZA==\"},\"https://ghcr.io\":{\"auth\":\"bWFuZGVsc29mdDp0b2tlbg==\"}},\"HttpHeaders\":{\"User-Agent\":\"Docker-Client/18.06.1-ce (linux)\"}},\"propagateConsumerIdentity\":true}"

		It("serializes repo spec", func() {
			configdata := Must(os.ReadFile("testdata/dockerconfig.json"))
			spec := local.NewRepositorySpecForConfig(configdata).WithConsumerPropagation(true)
			data := Must(json.Marshal(spec))

			var (
				datajson map[string]interface{}
				specjson map[string]interface{}
			)
			// Comparing the bytes might be problematic as the order of the JSON objects within the config file might change
			// during Marshaling
			MustBeSuccessful(json.Unmarshal([]byte(specdata), &specjson))
			MustBeSuccessful(json.Unmarshal(data, &datajson))
			Expect(datajson).To(Equal(specjson))
		})

		It("deserializes repo spec", func() {
			spec := Must(DefaultContext.RepositorySpecForConfig([]byte(specdata), nil))
			Expect(reflect.TypeOf(spec).String()).To(Equal("*dockerconfig.RepositorySpec"))
			configdata := Must(os.ReadFile("testdata/dockerconfig.json"))
			var (
				configdatajson   map[string]interface{}
				dockerconfigjson map[string]interface{}
			)
			// Comparing the bytes might be problematic as the order of the JSON objects within the config file might change
			// during Marshaling
			MustBeSuccessful(json.Unmarshal(configdata, &configdatajson))
			MustBeSuccessful(json.Unmarshal(spec.(*local.RepositorySpec).DockerConfig, &dockerconfigjson))
			Expect(dockerconfigjson).To(Equal(configdatajson))
		})

		It("resolves repository", func() {
			repo := Must(DefaultContext.RepositoryForConfig([]byte(specdata), nil))
			Expect(reflect.TypeOf(repo).String()).To(Equal("*dockerconfig.Repository"))
		})

		It("retrieves credentials", func() {
			repo := Must(DefaultContext.RepositoryForConfig([]byte(specdata), nil))

			creds := Must(repo.LookupCredentials("index.docker.io"))
			Expect(creds.Properties()).To(Equal(props))

			creds = Must(repo.LookupCredentials("ghcr.io"))
			Expect(creds.Properties()).To(Equal(props2))
		})

		It("propagates credentials to consumer identity", func() {
			Must(DefaultContext.RepositoryForConfig([]byte(specdata), nil))

			creds := Must(credentials.CredentialsForConsumer(DefaultContext, credentials.ConsumerIdentity{
				cpi.ATTR_TYPE:        identity.CONSUMER_TYPE,
				identity.ID_HOSTNAME: "ghcr.io",
			}))
			Expect(creds.Properties()).To(Equal(props2))
		})
	})

	Context("ref handling", func() {
		specdata := "{\"type\":\"DockerConfig\",\"dockerConfigFile\":\"testdata/dockerconfig.json\",\"propagateConsumerIdentity\":true}"

		It("can access the default context", func() {
			ctx := credentials.New()

			r := finalizer.GetRuntimeFinalizationRecorder(ctx)
			Expect(r).NotTo(BeNil())

			Must(ctx.RepositoryForConfig([]byte(specdata), nil))

			runtime.GC()
			time.Sleep(time.Second)
			ctx.GetType()
			Expect(r.Get()).To(BeNil())

			ctx = nil
			runtime.GC()
			time.Sleep(time.Second)

			Expect(r.Get()).To(ContainElement(ContainSubstring(credentials.CONTEXT_TYPE)))
		})
	})
})
