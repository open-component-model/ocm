package localfsblob_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/pkg/testutils"

	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/localfsblob"
	"github.com/open-component-model/ocm/pkg/mime"
)

var _ = Describe("Method", func() {

	var data = `globalAccess:
  digest: sha256:1bf729fa00e355199e711933ccfa27467ee3d2de1343aef2a7c1ecbdf885e63a
  mediaType: application/tar+gzip
  ref: ghcr.io/vasu1124/ocm/component-descriptors/github.com/vasu1124/introspect-delivery
  size: 11287
  type: ociBlob
localReference: sha256:1bf729fa00e355199e711933ccfa27467ee3d2de1343aef2a7c1ecbdf885e63a
mediaType: application/tar+gzip
type: localBlob
`
	_ = data

	It("marshal/unmarshal simple", func() {
		spec := localfsblob.New("path", mime.MIME_TEXT)
		data := Must(json.Marshal(spec))
		Expect(string(data)).To(Equal("{\"type\":\"localFilesystemBlob\",\"fileName\":\"path\",\"mediaType\":\"text/plain\"}"))
		r := Must(localfsblob.Decode(data))
		Expect(r).To(Equal(spec))
	})
})
