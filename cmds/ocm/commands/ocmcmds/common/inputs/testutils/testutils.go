package testutils

import (
	"encoding/json"

	"github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/gomega"
	"github.com/spf13/pflag"
	"ocm.software/ocm/api/utils/cobrautils/flagsets"
	"ocm.software/ocm/api/utils/runtime"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs"
)

type NameProvider interface {
	GetName() string
}

type InputTest struct {
	Type    inputs.InputType
	Options flagsets.ConfigOptions
	Flags   *pflag.FlagSet
}

func NewInputTest(name string) *InputTest {
	t := &InputTest{}
	t.Type = inputs.DefaultInputTypeScheme.GetInputType(name)
	t.Options = t.Type.ConfigOptionTypeSetHandler().CreateOptions()
	t.Flags = pflag.NewFlagSet("test", pflag.ContinueOnError)
	t.Options.AddFlags(t.Flags)
	return t
}

func (t *InputTest) Set(opt NameProvider, value string) {
	ExpectWithOffset(1, t.Flags.Set(opt.GetName(), value)).To(Succeed())
}

func (t *InputTest) SetWithFailure(opt NameProvider, value string, msg string) {
	err := t.Flags.Set(opt.GetName(), value)
	ExpectWithOffset(1, err).To(HaveOccurred())
	ExpectWithOffset(1, err.Error()).To(Equal(msg))
}

func (t *InputTest) Check(expected interface{}) {
	config := flagsets.Config{}
	ExpectWithOffset(1, t.Type.ConfigOptionTypeSetHandler().ApplyConfig(t.Options, config)).To(Succeed())
	data, err := json.Marshal(config)
	ExpectWithOffset(1, err).To(Succeed())
	spec, err := t.Type.Decode(data, runtime.DefaultJSONEncoding)
	ExpectWithOffset(1, err).To(Succeed())
	ExpectWithOffset(1, spec).To(testutils.DeepEqual(expected))
}
