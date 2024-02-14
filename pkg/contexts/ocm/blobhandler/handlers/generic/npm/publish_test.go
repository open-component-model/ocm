// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package npm

import (
	"encoding/json"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/open-component-model/ocm/pkg/testutils"
)

var _ = Describe("NPM Publish Test Environment", func() {

	It("prepare package and json marshal", func() {
		data := Must(os.ReadFile("testdata/testdata.tgz"))
		pkg := Must(prepare(data))
		Expect(pkg.Version).To(Equal("0.8.15"))
		Expect(pkg.Name).To(Equal("testdata"))
		Expect(pkg.Dist.Integrity).To(Equal("sha512-Dsbnf3b4scCugxBZ+rHm8Hr1CAfyC3h8su31KnPGw21BAkM6X5gbi5Jbci9WaCCBBxm1tMTRKCJqk29j5Aw0gg=="))
		jsn := Must(json.Marshal(pkg))
		fixture := `{"name":"testdata","version":"0.8.15","readme":"# Test Data\n\nreadme\n","description":"Test data description.","dist":{"integrity":"sha512-Dsbnf3b4scCugxBZ+rHm8Hr1CAfyC3h8su31KnPGw21BAkM6X5gbi5Jbci9WaCCBBxm1tMTRKCJqk29j5Aw0gg==","shasum":"602b69a43903fa1694d59fefd8cb326cd68e8935","tarball":""}}`
		Expect(string(jsn)).To(Equal(fixture))
	})

})
