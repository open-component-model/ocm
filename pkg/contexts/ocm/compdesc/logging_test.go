// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package compdesc_test

import (
	"bytes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/go-logr/logr"
	"github.com/mandelsoft/logging"
	"github.com/tonglil/buflogr"

	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	ocmlog "github.com/open-component-model/ocm/pkg/logging"
	"github.com/open-component-model/ocm/pkg/testutils"
)

var _ = Describe("logging", func() {
	var old logr.Logger
	var oldlevel int
	var buf *bytes.Buffer

	BeforeEach(func() {
		buf = bytes.NewBuffer(nil)
		log := buflogr.NewWithBuffer(buf)
		old = logr.New(ocmlog.Context().GetSink())
		oldlevel = ocmlog.Context().GetDefaultLevel()
		ocmlog.Context().SetBaseLogger(log)
		ocmlog.Context().SetDefaultLevel(logging.DebugLevel)
	})

	AfterEach(func() {
		ocmlog.Context().SetBaseLogger(old)
		ocmlog.Context().SetDefaultLevel(oldlevel)
	})

	It("logs failures", func() {

		_, err := compdesc.Decode([]byte("[]"))
		Expect(err).To(MatchError(`error unmarshaling JSON: while decoding JSON: json: cannot unmarshal array into Go value of type struct { Meta v1.Metadata "json:\"meta\""; APIVersion string "json:\"apiVersion\"" }`))
		Expect(buf.String()).To(testutils.StringEqualTrimmedWithContext(`
V[4] ocm/compdesc decoding of component descriptor failed error error unmarshaling JSON: while decoding JSON: json: cannot unmarshal array into Go value of type struct { Meta v1.Metadata "json:\"meta\""; APIVersion string "json:\"apiVersion\"" } data []
`))
	})

	It("logs format failures", func() {

		_, err := compdesc.Decode([]byte(`
meta:
  schemaVersion: v2
component:
  name: acme.org/test
  version: 1.0.0
  provider: acme.org 
  creationTime: "0815"
  repositoryContexts: []
  resources: []
  sources: []
  componentReferences: []
`))
		Expect(err).To(MatchError(`component.creationTime: Does not match format 'date-time'`))
		Expect(buf.String()).To(testutils.StringEqualTrimmedWithContext(`
V[4] ocm/compdesc versioned decoding of component descriptor failed error component.creationTime: Does not match format 'date-time' scheme v2 data meta: schemaVersion: v2 component: name: acme.org/test version: 1.0.0 provider: acme.org creationTime: "0815" repositoryContexts: [] resources: [] sources: [] componentReferences: []
`))
	})

})
