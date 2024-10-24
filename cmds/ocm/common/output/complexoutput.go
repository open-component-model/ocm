package output

import (
	"fmt"

	. "ocm.software/ocm/api/utils/out"
	. "ocm.software/ocm/cmds/ocm/common/processing"

	"sigs.k8s.io/yaml"

	"ocm.software/ocm/api/utils"
	"ocm.software/ocm/api/utils/runtime"
)

type ComplexProcessingOutput struct {
	ElementOutput
	mapper func(interface{}) interface{}
	fields []string
}

var _ Output = &ComplexProcessingOutput{}

func NewProcessingComplexOutput(opts *Options, chain ProcessChain, fields ...string) *ComplexProcessingOutput {
	return (&ComplexProcessingOutput{}).new(opts, chain, fields)
}

func (this *ComplexProcessingOutput) new(opts *Options, chain ProcessChain, fields []string) *ComplexProcessingOutput {
	this.ElementOutput.new(opts, chain)
	this.fields = fields
	return this
}

func (this *ComplexProcessingOutput) Out() error {
	i := this.Elems.Iterator()
	for i.HasNext() {
		Outf(this.Context, "---\n")
		elem := i.Next()
		var out interface{}
		if this.mapper != nil {
			out = this.mapper(elem)
		}
		data, err := runtime.DefaultYAMLEncoding.Marshal(out)
		if err != nil {
			Error(this.Context, err.Error())
		} else {
			if len(this.fields) > 0 {
				m := map[string]interface{}{}
				runtime.DefaultYAMLEncoding.Unmarshal(data, m)
				this.out("", m)
			} else {
				Outf(this.Context, "%s\n", string(data))
			}
		}
	}
	return this.ElementOutput.Out()
}

func (this *ComplexProcessingOutput) out(gap string, m map[string]interface{}) {
	rest := map[string]interface{}{}
	for k, v := range m {
		rest[k] = v
	}

	for _, k := range this.fields {
		v := m[k]
		delete(rest, k)
		if v != nil {
			switch e := v.(type) {
			case map[string]interface{}:
				Outf(this.Context, "%s%s:\n", gap, k)
				this.out(gap+"  ", e)
			case []interface{}:
				Outf(this.Context, "%s%s:\n", gap, k)
				s, err := yaml.Marshal(v)
				if err == nil {
					utils.IndentLines(string(s), gap)
				}
			default:
				eff := utils.IndentLines(fmt.Sprintf("%v", v), gap+"  ")
				Outf(this.Context, "%s%s: %s", gap, k, eff[len(gap)+2:])
			}
		}
	}
	s, err := yaml.Marshal(rest)
	if err == nil {
		Out(this.Context, utils.IndentLines(string(s), gap))
	}
}
