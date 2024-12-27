package transferhandler_test

import (
	"reflect"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"ocm.software/ocm/api/ocm/tools/transfer/transferhandler"

	_ "ocm.software/ocm/api/ocm/tools/transfer"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/generics"

	"ocm.software/ocm/api/datacontext"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/plugin"
	"ocm.software/ocm/api/ocm/tools/transfer/transferhandler/spiff"
	"ocm.software/ocm/api/ocm/tools/transfer/transferhandler/standard"
)

var _ = Describe("Registration Test Environment", func() {
	ctx := ocm.New(datacontext.MODE_EXTENDED)

	It("standard", func() {
		h := Must(transferhandler.For(ctx).ByName(ctx, "ocm/standard"))
		Expect(reflect.TypeOf(h)).To(Equal(generics.TypeOf[*standard.Handler]()))
	})

	It("spiff", func() {
		h := Must(transferhandler.For(ctx).ByName(ctx, "ocm/spiff"))
		Expect(reflect.TypeOf(h)).To(Equal(generics.TypeOf[*spiff.Handler]()))
	})

	It("plugin", func() {
		ExpectError(transferhandler.For(ctx).ByName(ctx, "plugin/p/h")).To(MatchError(errors.ErrUnknown(plugin.KIND_PLUGIN, "p")))
	})
})
