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

package dockerconfig_test

import (
	"context"
	"encoding/json"
	"reflect"

	"github.com/gardener/ocm/pkg/common"
	"github.com/gardener/ocm/pkg/credentials"
	local "github.com/gardener/ocm/pkg/credentials/repositories/dockerconfig"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var DefaultContext = credentials.NewDefaultContext(context.TODO())

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

	specdata := "{\"type\":\"DockerConfig\",\"dockerConfigFile\":\"testdata/dockerconfig.json\"}"

	_ = props

	It("serializes repo spec", func() {
		spec := local.NewRepositorySpec("testdata/dockerconfig.json")
		data, err := json.Marshal(spec)
		Expect(err).To(Succeed())
		Expect(data).To(Equal([]byte(specdata)))
	})
	It("deserializes repo spec", func() {
		spec, err := DefaultContext.RepositorySpecForConfig([]byte(specdata), nil)
		Expect(err).To(Succeed())
		Expect(reflect.TypeOf(spec).String()).To(Equal("*dockerconfig.RepositorySpec"))
		Expect(spec.(*local.RepositorySpec).DockerConfigFile).To(Equal("testdata/dockerconfig.json"))
	})

	It("resolves repository", func() {
		repo, err := DefaultContext.RepositoryForConfig([]byte(specdata), nil)
		Expect(err).To(Succeed())
		Expect(reflect.TypeOf(repo).String()).To(Equal("*dockerconfig.Repository"))
	})

	It("retrieves credentials", func() {
		repo, err := DefaultContext.RepositoryForConfig([]byte(specdata), nil)
		Expect(err).To(Succeed())

		creds, err := repo.LookupCredentials("index.docker.io")
		Expect(err).To(Succeed())
		Expect(creds.Properties()).To(Equal(props))

		creds, err = repo.LookupCredentials("ghcr.io")
		Expect(err).To(Succeed())
		Expect(creds.Properties()).To(Equal(props2))
	})
})
