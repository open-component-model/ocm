// Copyright 2022 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package config_test

import (
	"encoding/json"
	"reflect"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/config"
	"github.com/open-component-model/ocm/pkg/credentials"
	localconfig "github.com/open-component-model/ocm/pkg/credentials/config"
	"github.com/open-component-model/ocm/pkg/credentials/repositories/aliases"
	"github.com/open-component-model/ocm/pkg/credentials/repositories/directcreds"
	"github.com/open-component-model/ocm/pkg/credentials/repositories/memory"
)

var DefaultContext = credentials.New()

var _ = Describe("generic credentials", func() {
	props := common.Properties{
		"user":     "USER",
		"password": "PASSWORD",
	}

	repospec := memory.NewRepositorySpec("test")
	credspec := credentials.NewCredentialsSpec("cred", repospec)
	direct := directcreds.NewRepositorySpec(props)

	cfgconsumerdata := "{\"type\":\"credentials.config.ocm.gardener.cloud\",\"consumers\":[{\"identity\":{\"type\":\"oci\",\"url\":\"https://acme.com\"},\"credentials\":[{\"credentialsName\":\"cred\",\"repoName\":\"test\",\"type\":\"Memory\"}]}]}"
	cfgrepodata := "{\"type\":\"credentials.config.ocm.gardener.cloud\",\"repositories\":[{\"repository\":{\"repoName\":\"test\",\"type\":\"Memory\"},\"credentials\":[{\"properties\":{\"password\":\"PASSWORD\",\"user\":\"USER\"},\"type\":\"Credentials\"}]}]}"
	cfgaliasdata := "{\"type\":\"credentials.config.ocm.gardener.cloud\",\"aliases\":{\"alias\":{\"repository\":{\"repoName\":\"test\",\"type\":\"Memory\"},\"credentials\":[{\"properties\":{\"password\":\"PASSWORD\",\"user\":\"USER\"},\"type\":\"Credentials\"}]}}}"
	_ = props

	Context("serialize", func() {

		It("serializes repository spec not in map", func() {
			mapdata := "{\"repositories\":{\"repository\":{\"repoName\":\"test\",\"type\":\"Memory\"}}}"
			type S struct {
				Repositories localconfig.RepositorySpec `json:"repositories"`
			}

			rspec, err := credentials.ToGenericRepositorySpec(repospec)
			Expect(err).To(Succeed())
			s := &S{
				Repositories: localconfig.RepositorySpec{Repository: *rspec},
			}
			data, err := json.Marshal(s)

			Expect(err).To(Succeed())
			Expect(data).To(Equal([]byte(mapdata)))
		})

		It("serializes repository spec map", func() {
			mapdata := "{\"repositories\":{\"repo\":{\"repository\":{\"repoName\":\"test\",\"type\":\"Memory\"}}}}"
			type S struct {
				Repositories map[string]localconfig.RepositorySpec `json:"repositories"`
			}

			rspec, err := credentials.ToGenericRepositorySpec(repospec)
			Expect(err).To(Succeed())
			s := &S{
				Repositories: map[string]localconfig.RepositorySpec{
					"repo": localconfig.RepositorySpec{Repository: *rspec},
				},
			}
			data, err := json.Marshal(s)
			Expect(err).To(Succeed())
			Expect(data).To(Equal([]byte(mapdata)))
		})
	})

	Context("composition", func() {
		It("composes a config for consumers", func() {
			consumerid := credentials.ConsumerIdentity{
				"type": "oci",
				"url":  "https://acme.com",
			}

			cfg := localconfig.NewConfigSpec()

			cfg.AddConsumer(consumerid, credspec)

			data, err := json.Marshal(cfg)
			Expect(err).To(Succeed())
			Expect(data).To(Equal([]byte(cfgconsumerdata)))

			cfg2 := &localconfig.ConfigSpec{}
			err = json.Unmarshal(data, cfg2)
			Expect(err).To(Succeed())
			Expect(cfg2).To(Equal(cfg))
		})

		It("composes a config for repositories", func() {

			cfg := localconfig.NewConfigSpec()

			cfg.AddRepository(repospec, direct)

			data, err := json.Marshal(cfg)
			Expect(err).To(Succeed())
			Expect(data).To(Equal([]byte(cfgrepodata)))

			cfg2 := &localconfig.ConfigSpec{}
			err = json.Unmarshal(data, cfg2)
			Expect(err).To(Succeed())
			Expect(cfg2).To(Equal(cfg))
		})

		It("composes a config for aliases", func() {

			cfg := localconfig.NewConfigSpec()

			cfg.AddAlias("alias", repospec, direct)

			data, err := json.Marshal(cfg)
			Expect(err).To(Succeed())
			Expect(data).To(Equal([]byte(cfgaliasdata)))

			cfg2 := &localconfig.ConfigSpec{}
			err = json.Unmarshal(data, cfg2)
			Expect(err).To(Succeed())
			Expect(cfg2).To(Equal(cfg))
		})
	})

	Context("apply", func() {
		var ctx credentials.Context

		_ = ctx

		BeforeEach(func() {
			ctx = credentials.WithConfigs(config.New()).New()
		})

		It("applies a config for aliases", func() {
			cfg := localconfig.NewConfigSpec()
			cfg.AddAlias("alias", repospec, direct)

			ctx.ConfigContext().ApplyConfig(cfg, "testconfig")

			spec := aliases.NewRepositorySpec("alias")

			repo, err := ctx.RepositoryForSpec(spec)
			Expect(err).To(Succeed())
			Expect(reflect.TypeOf(repo).String()).To(Equal("*memory.Repository"))
		})
	})
})
