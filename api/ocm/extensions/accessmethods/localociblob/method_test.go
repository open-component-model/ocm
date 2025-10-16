package localociblob_test

import (
	"encoding/json"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/localociblob"
)

var _ = Describe("Method", func() {
	It("marshal/unmarshal simple", func() {
		spec := localociblob.New("sha256:1bf729fa00e355199e711933ccfa27467ee3d2de1343aef2a7c1ecbdf885e63a")
		data := Must(json.Marshal(spec))
		Expect(string(data)).To(Equal("{\"type\":\"localOciBlob\",\"digest\":\"sha256:1bf729fa00e355199e711933ccfa27467ee3d2de1343aef2a7c1ecbdf885e63a\"}"))
		r := Must(localociblob.Decode(data))
		Expect(r).To(Equal(spec))
	})
})
