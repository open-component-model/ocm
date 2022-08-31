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

package directcreds_test

import (
	"encoding/json"
	"reflect"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	"github.com/open-component-model/ocm/pkg/contexts/credentials/repositories/directcreds"
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
