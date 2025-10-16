package flagsets_test

import (
	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/pflag"
	"ocm.software/ocm/api/utils/cobrautils/flagsets"
)

func adder(opts flagsets.ConfigOptions, data flagsets.Config) error {
	for _, o := range opts.Options() {
		if o.Changed() {
			data[o.GetName()] = o.Value()
		}
	}
	return nil
}

var _ = Describe("configure", func() {
	var sett1 flagsets.ConfigOptionTypeSet
	var sett2 flagsets.ConfigOptionTypeSet

	var prov flagsets.ConfigTypeOptionSetConfigProvider
	var flags *pflag.FlagSet
	var opts flagsets.ConfigOptions

	BeforeEach(func() {
		sett1 = flagsets.NewPlainConfigProvider("t1", adder)
		sett1.AddOptionType(flagsets.NewStringOptionType("common", "a test string"))
		sett1.AddOptionType(flagsets.NewStringOptionType("t1", "t1 argument"))

		sett2 = flagsets.NewPlainConfigProvider("t2", adder)
		sett2.AddOptionType(flagsets.NewStringOptionType("common", "a test string"))
		sett2.AddOptionType(flagsets.NewStringOptionType("t2", "t2 argument"))

		prov = flagsets.NewTypedConfigProvider("entry", "entry data", "entryType")
		prov.AddTypeSet(sett1)
		prov.AddTypeSet(sett2)
		flags = pflag.NewFlagSet("flags", pflag.ContinueOnError)
		opts = prov.CreateOptions()

		opts.AddFlags(flags)
	})

	It("fills entry", func() {
		Expect(flags.Parse([]string{"--entryType=t1", "--common=string", "--t1=value"})).To(Succeed())

		data := Must(prov.GetConfigFor(opts))
		Expect(data).To(Equal(flagsets.Config{
			"type":   "t1",
			"common": "string",
			"t1":     "value",
		}))
	})

	It("detects non-matching option", func() {
		Expect(flags.Parse([]string{"--entryType=t1", "--common=string", "--t1=value", "--t2=value"})).To(Succeed())
		ExpectError(prov.GetConfigFor(opts)).To(MatchError(`option "t2" given, but not possible for entry type t1`))
	})
})
