package valuesets

import (
	out "fmt"
	"strings"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/set"

	"ocm.software/ocm/api/ocm/extensions/accessmethods/options"
	"ocm.software/ocm/api/ocm/plugin/descriptor"
	"ocm.software/ocm/api/ocm/plugin/ppi"
	"ocm.software/ocm/api/utils"
	"ocm.software/ocm/api/utils/cobrautils/flagsets"
	"ocm.software/ocm/api/utils/runtime"
)

const NAME = "check"

type Value struct {
	runtime.ObjectTypedObject `json:",inline"`
	Checks                    map[string]Status `json:"checks"`
}

type Status struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

const (
	STATUS_PASSED  = "passed"
	STATUS_FAILED  = "failed"
	STATUS_SKIPPED = "skipped"
)

var status = set.New[string](STATUS_PASSED, STATUS_FAILED, STATUS_SKIPPED)

var (
	StatusOption  = options.NewStringMapOptionType("checkStatus", out.Sprintf("status value for check (%s)", strings.Join(utils.StringMapKeys(status), ", ")))
	MessageOption = options.NewStringMapOptionType("checkMessage", "message for check")
)

type ValueSet struct {
	ppi.ValueSetBase
}

func New() ppi.ValueSet {
	return &ValueSet{
		ValueSetBase: ppi.MustNewValueSetBase(NAME, "", &Value{}, []string{descriptor.PURPOSE_ROUTINGSLIP}, "set of check status", `
- **<code>checks</code>** *map{string]status* set of status entries

The status entry has the following format:
- **<code>status</code> *string* status code (passed, failed)
- **<code>message</code> *string* message
`),
	}
}

func (v ValueSet) Options() []options.OptionType {
	return []options.OptionType{
		StatusOption,
		MessageOption,
	}
}

func (v ValueSet) ValidateSpecification(_ ppi.Plugin, spec runtime.TypedObject) (*ppi.ValueSetInfo, error) {
	var info ppi.ValueSetInfo

	my := spec.(*Value)

	desc := ""
	for c, v := range my.Checks {
		if v.Status == "" {
			return nil, out.Errorf("status not specified")
		}
		if !status.Contains(v.Status) {
			return nil, out.Errorf("invalid status (%s), expected %s", v.Status, strings.Join(utils.StringMapKeys(status), ", "))
		}

		if len(desc) > 0 {
			desc += ", "
		}
		desc += c + ": " + v.Status
	}

	info.Short = desc
	return &info, nil
}

func (v ValueSet) ComposeSpecification(_ ppi.Plugin, opts ppi.Config, config ppi.Config) error {
	list := errors.ErrListf("configuring options")

	if v, ok := opts.GetValue(StatusOption.GetName()); ok {
		for c, s := range v.(map[string]string) {
			list.Addf(nil, flagsets.SetField(config, s, "checks", c, "status"), "status")
		}
	}
	if v, ok := opts.GetValue(MessageOption.GetName()); ok {
		for c, s := range v.(map[string]string) {
			list.Addf(nil, flagsets.SetField(config, s, "checks", c, "message"), "message")
		}
	}

	return list.Result()
}
