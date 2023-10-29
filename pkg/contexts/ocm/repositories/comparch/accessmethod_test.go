// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package comparch_test

import (
	"encoding/json"
	"os"
	"reflect"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/localblob"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/localfsblob"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi/accspeccpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/comparch"
	"github.com/open-component-model/ocm/pkg/runtime"
)

var DefaultContext = ocm.New()

var _ = Describe("access method", func() {

	legacy := "{\"type\":\"localFilesystemBlob\",\"fileName\":\"anydigest\",\"mediaType\":\"application/json\"}"

	Context("local access method", func() {
		It("decodes legacy methood", func() {
			spec, err := DefaultContext.AccessSpecForConfig([]byte(legacy), nil)
			Expect(err).To(Succeed())
			Expect(reflect.TypeOf(spec)).To(Equal(reflect.TypeOf(&localblob.AccessSpec{})))
			Expect(spec.(*localblob.AccessSpec).LocalReference).To(Equal("anydigest"))
		})

		It("encodes legacy methood", func() {
			spec := localfsblob.New("anydigest", "application/json")
			data, err := DefaultContext.Encode(spec, runtime.DefaultJSONEncoding)
			Expect(err).To(Succeed())
			Expect(data).To(Equal([]byte(legacy)))
		})
	})

	Context("component archive", func() {
		It("instantiate local blob access method for component archive", func() {
			data, err := os.ReadFile("testdata/descriptor/component-descriptor.yaml")
			Expect(err).To(Succeed())
			cd, err := compdesc.Decode(data)
			Expect(err).To(Succeed())

			ca, err := comparch.New(DefaultContext, accessobj.ACC_CREATE, nil, nil, nil, 0600)
			Expect(err).To(Succeed())

			ca.GetDescriptor().Name = "acme.org/dummy"
			ca.GetDescriptor().Version = "v1"

			res, err := cd.GetResourceByIdentity(metav1.NewIdentity("local"))
			Expect(err).To(Succeed())
			Expect(res).To(Not(BeNil()))

			spec, err := DefaultContext.AccessSpecForSpec(res.Access)
			Expect(err).To(Succeed())
			Expect(spec).To(Not(BeNil()))

			Expect(spec.GetType()).To(Equal(localfsblob.Type))
			Expect(spec.GetKind()).To(Equal(localfsblob.Type))
			Expect(spec.GetVersion()).To(Equal("v1"))
			Expect(reflect.TypeOf(spec)).To(Equal(reflect.TypeOf(&localblob.AccessSpec{})))

			data, err = json.Marshal(spec)
			Expect(err).To(Succeed())
			Expect(string(data)).To(Equal(legacy))

			m, err := spec.AccessMethod(ca)
			Expect(err).To(Succeed())
			Expect(m).To(Not(BeNil()))
			Expect(reflect.TypeOf(accspeccpi.GetAccessMethodImplementation(m)).String()).To(Equal("*comparch.localFilesystemBlobAccessMethod"))
			Expect(m.GetKind()).To(Equal("localBlob"))

			Expect(ca.Close()).To(Succeed())
		})
	})
})
