package transfer_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/pkg/testutils"

	"github.com/mandelsoft/goutils/errors"

	"github.com/open-component-model/ocm/pkg/contexts/ocm/transfer"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/transfer/transferhandler"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/transfer/transferhandler/spiff"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/transfer/transferhandler/standard"
)

type errorOption struct {
	standard.TransferOptionsCreator
}

var _ transferhandler.TransferOption = (*errorOption)(nil)

func (e *errorOption) ApplyTransferOption(options transferhandler.TransferOptions) error {
	return errors.ErrNotSupported(transferhandler.KIND_TRANSFEROPTION, "error")
}

var _ = Describe("auto detection", func() {
	It("detects default handler", func() {
		h := Must(transfer.NewTransferHandler())

		n := Must(standard.New())
		Expect(h).To(Equal(n))
	})

	It("detects standard handler", func() {
		h := Must(transfer.NewTransferHandler(standard.Recursive()))

		n := Must(standard.New(standard.Recursive()))
		Expect(h).To(Equal(n))
	})

	It("detects spiff handler", func() {
		h := Must(transfer.NewTransferHandler(spiff.Script([]byte(""))))

		n := Must(spiff.New())
		Expect(h).To(Equal(n))
	})

	It("detects spiff handler for leading standard options", func() {
		h := Must(transfer.NewTransferHandler(standard.Recursive(), spiff.Script([]byte(""))))

		n := Must(spiff.New(standard.Recursive()))
		Expect(h).To(Equal(n))
	})

	It("fails on invalid option", func() {
		_, err := transfer.NewTransferHandler(&errorOption{}, standard.Recursive(), spiff.Script([]byte("")))

		Expect(err).To(Equal(errors.ErrNotSupported(transferhandler.KIND_TRANSFEROPTION, "error")))
	})

})
