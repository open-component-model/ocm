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

package core_test

import (
	"encoding/json"
	"reflect"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/open-component-model/ocm/pkg/common"
	credentials2 "github.com/open-component-model/ocm/pkg/contexts/credentials"
	core2 "github.com/open-component-model/ocm/pkg/contexts/credentials/core"
	"github.com/open-component-model/ocm/pkg/contexts/credentials/repositories/memory"
)

var DefaultContext = credentials2.New()

var _ = Describe("generic credentials", func() {
	props := common.Properties{
		"user":     "USER",
		"password": "PASSWORD",
	}
	credmemdata := "{\"credentialsName\":\"cred\",\"repoName\":\"test\",\"type\":\"Memory\"}"
	memdata := "{\"repoName\":\"test\",\"type\":\"Memory\"}"

	_ = props

	It("de/serializes credentials spec", func() {
		repospec := memory.NewRepositorySpec("test")
		credspec := credentials2.NewCredentialsSpec("cred", repospec)

		data, err := json.Marshal(credspec)
		Expect(err).To(Succeed())
		Expect(data).To(Equal([]byte(credmemdata)))

		credspec = &core2.DefaultCredentialsSpec{}
		err = json.Unmarshal(data, credspec)
		Expect(err).To(Succeed())
		s := credspec.(*core2.DefaultCredentialsSpec)
		Expect(reflect.TypeOf(s.RepositorySpec).String()).To(Equal("*memory.RepositorySpec"))
		Expect(s.CredentialsName).To(Equal("cred"))
		Expect(s.RepositorySpec.(*memory.RepositorySpec).RepositoryName).To(Equal("test"))
	})

	It("de/serializes generic credentials spec", func() {
		credspec := &core2.GenericCredentialsSpec{}

		err := json.Unmarshal([]byte(credmemdata), credspec)
		Expect(err).To(Succeed())

		data, err := json.Marshal(credspec)
		Expect(err).To(Succeed())
		Expect(data).To(Equal([]byte(credmemdata)))
	})

	It("de/serializes generic repository spec", func() {
		credspec := &core2.GenericRepositorySpec{}

		err := json.Unmarshal([]byte(memdata), credspec)
		Expect(err).To(Succeed())

		data, err := json.Marshal(credspec)
		Expect(err).To(Succeed())
		Expect(data).To(Equal([]byte(memdata)))
	})

	It("converts credentials spec to generic ones", func() {
		repospec := memory.NewRepositorySpec("test")
		credspec := credentials2.NewCredentialsSpec("cred", repospec)
		data, err := json.Marshal(credspec)
		Expect(err).To(Succeed())

		gen, err := credentials2.ToGenericCredentialsSpec(credspec)
		Expect(err).To(Succeed())

		Expect(reflect.TypeOf(gen).String()).To(Equal("*core.GenericCredentialsSpec"))
		Expect(reflect.TypeOf(gen.RepositorySpec).String()).To(Equal("*core.GenericRepositorySpec"))

		gen2, err := credentials2.ToGenericCredentialsSpec(gen)
		Expect(err).To(Succeed())
		Expect(gen2).To(BeIdenticalTo(gen))

		data3, err := json.Marshal(gen)
		Expect(err).To(Succeed())
		Expect(data3).To(Equal(data))
	})
})
