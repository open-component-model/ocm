// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

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
	It("marshal/unmarshal simple", func() {
		spec := localfsblob.New("path", mime.MIME_TEXT)
		data := Must(json.Marshal(spec))
		Expect(string(data)).To(Equal("{\"type\":\"localFilesystemBlob\",\"filename\":\"path\",\"mediaType\":\"text/plain\"}"))
		r := Must(localfsblob.Decode(data))
		Expect(r).To(Equal(spec))
	})
})
