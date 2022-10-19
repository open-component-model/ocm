// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package aliases_test

import (
	"encoding/json"
	"reflect"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	local "github.com/open-component-model/ocm/pkg/contexts/credentials/repositories/aliases"
)

var DefaultContext = credentials.New()

var _ = Describe("alias credentials", func() {

	props := common.Properties{
		"user":     "USER",
		"password": "PASSWORD",
	}

	memorydata := "{\"type\":\"Memory\",\"repoName\":\"myrepo\"}"
	specdata := "{\"type\":\"Alias\",\"alias\":\"test\"}"

	It("serializes repo spec", func() {
		spec := local.NewRepositorySpec("test")
		data, err := json.Marshal(spec)
		Expect(err).To(Succeed())
		Expect(data).To(Equal([]byte(specdata)))
	})
	It("deserializes repo spec", func() {
		spec, err := DefaultContext.RepositorySpecForConfig([]byte(specdata), nil)
		Expect(err).To(Succeed())
		Expect(reflect.TypeOf(spec).String()).To(Equal("*aliases.RepositorySpec"))
		Expect(spec.(*local.RepositorySpec).Alias).To(Equal("test"))
	})

	It("resolves repository", func() {
		memoryspec, err := credentials.NewGenericRepositorySpec([]byte(memorydata), nil)
		Expect(err).To(Succeed())

		err = DefaultContext.SetAlias("test", memoryspec)
		Expect(err).To(Succeed())

		repo, err := DefaultContext.RepositoryForConfig([]byte(specdata), nil)
		Expect(err).To(Succeed())
		Expect(reflect.TypeOf(repo).String()).To(Equal("*memory.Repository"))
	})

	It("sets and retrieves credentials", func() {
		memoryspec, err := credentials.NewGenericRepositorySpec([]byte(memorydata), nil)
		Expect(err).To(Succeed())

		err = DefaultContext.SetAlias("test", memoryspec)
		Expect(err).To(Succeed())

		repo, err := DefaultContext.RepositoryForConfig([]byte(memorydata), nil)
		Expect(err).To(Succeed())

		_, err = repo.WriteCredentials("bibo", credentials.NewCredentials(props))
		Expect(err).To(Succeed())

		credspec := credentials.NewCredentialsSpec("bibo", local.NewRepositorySpec("test"))

		creds, err := DefaultContext.CredentialsForSpec(credspec)
		Expect(err).To(Succeed())
		Expect(creds.Properties()).To(Equal(props))
	})

})
