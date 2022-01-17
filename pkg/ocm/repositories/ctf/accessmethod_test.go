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

package ctf_test

import (
	"context"
	"encoding/json"
	"os"
	"reflect"

	_ "github.com/gardener/ocm/pkg/ocm"
	"github.com/gardener/ocm/pkg/ocm/accessmethods"
	"github.com/gardener/ocm/pkg/ocm/compdesc"
	metav1 "github.com/gardener/ocm/pkg/ocm/compdesc/meta/v1"
	"github.com/gardener/ocm/pkg/ocm/core"
	"github.com/gardener/ocm/pkg/ocm/repositories/ctf"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var DefaultContext = core.NewDefaultContext(context.TODO())

var _ = Describe("access method", func() {

	It("instantiate local blob access method for component archive", func() {
		data, err := os.ReadFile("testdata/component-descriptor.yaml")
		Expect(err).To(Succeed())
		cd, err := compdesc.Decode(data)
		Expect(err).To(Succeed())

		ca := ctf.NewComponentArchive(DefaultContext, nil, cd, nil, nil)

		res, err := cd.GetResourceByIdentity(metav1.IdentityByName("local"))
		Expect(err).To(Succeed())
		Expect(res).To(Not(BeNil()))

		spec, err := DefaultContext.AccessSpecForSpec(res.Access)
		Expect(err).To(Succeed())
		Expect(spec).To(Not(BeNil()))

		Expect(spec.GetType()).To(Equal(accessmethods.LocalBlobType))
		Expect(spec.GetName()).To(Equal(accessmethods.LocalBlobType))
		Expect(spec.GetVersion()).To(Equal("v1"))
		Expect(reflect.TypeOf(spec).String()).To(Equal("*accessmethods.LocalBlobAccessSpec"))

		data, err = json.Marshal(spec)
		Expect(err).To(Succeed())
		Expect(string(data)).To(Equal("{\"type\":\"localBlob\",\"filename\":\"\",\"mediaType\":\"application/json\"}"))

		m, err := spec.AccessMethod(ca)
		Expect(err).To(Succeed())
		Expect(m).To(Not(BeNil()))
		Expect(reflect.TypeOf(m).String()).To(Equal("*ctf.LocalFilesystemBlobAccessMethod"))
	})
})
