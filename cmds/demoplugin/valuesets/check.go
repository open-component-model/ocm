// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package valuesets

import (
	out "fmt"

	"github.com/open-component-model/ocm/pkg/cobrautils/flagsets"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/options"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/descriptor"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/ppi"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/runtime"
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
	STATUS_PASSED = "passed"
	STATUS_FAILED = "failed"
)

var (
	StatusOption  = options.NewStringMapOptionType("checkStatus", "status value for check")
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
- **<code>message</code> *string* mwssage
`),
	}
}

func (v ValueSet) Options() []options.OptionType {
	return []options.OptionType{
		StatusOption,
		MessageOption,
	}
}

func (v ValueSet) ValidateSpecification(p ppi.Plugin, spec runtime.TypedObject) (*ppi.ValueSetInfo, error) {
	var info ppi.ValueSetInfo

	my := spec.(*Value)

	desc := ""
	for c, v := range my.Checks {
		if v.Status == "" {
			return nil, out.Errorf("status not specified")
		}
		if v.Status != STATUS_FAILED && v.Status != STATUS_PASSED {
			return nil, out.Errorf("invalid status (%s), expectd %s or %s", v.Status, STATUS_PASSED, STATUS_FAILED)
		}

		if len(desc) > 0 {
			desc += ", "
		}
		desc += c + ": " + v.Status
	}

	info.Short = desc
	return &info, nil
}

func (v ValueSet) ComposeSpecification(p ppi.Plugin, opts ppi.Config, config ppi.Config) error {
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
