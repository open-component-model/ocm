package standard_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mandelsoft/goutils/generics"

	me "ocm.software/ocm/api/ocm/tools/transfer/transferhandler/standard"
)

type Target struct {
	flag *bool
	text string
}

//////

var _ Flag = (*Target)(nil)

type Flag interface {
	SetFlag(b bool)
}

func (t *Target) SetFlag(b bool) {
	t.flag = &b
}

//////

var _ Text = (*Target)(nil)

type Text interface {
	SetText(v string)
}

func (t *Target) SetText(v string) {
	t.text = v
}

var _ = Describe("utils", func() {
	It("handles pointer arg", func() {
		var v *bool
		t := &Target{}

		me.HandleOption[Flag](v, t)
		Expect(t.flag).To(BeNil())

		v = generics.PointerTo(false)
		me.HandleOption[Flag](v, t)
		Expect(t.flag).NotTo(BeNil())
		Expect(*t.flag).To(BeFalse())
	})

	It("handles value arg", func() {
		var v string
		t := &Target{text: "old"}

		me.HandleOption[Text](v, t)
		Expect(t.text).To(Equal("old"))

		v = "test"
		me.HandleOption[Text](v, t)
		Expect(t.text).To(Equal("test"))
	})
})
