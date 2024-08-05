package add

import (
	"encoding/json"

	"ocm.software/ocm/api/utils/cobrautils/flagsets"
	ocmcomm "ocm.software/ocm/cmds/ocm/commands/ocmcmds/common"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/addhdlrs"
)

type ReferenceResourceSpecificationProvider struct {
	*ocmcomm.ElementMetaDataSpecificationsProvider
}

var (
	_ ocmcomm.ElementSpecificationsProvider = (*ReferenceResourceSpecificationProvider)(nil)
	_ addhdlrs.ElementSource                = (*ReferenceResourceSpecificationProvider)(nil)
)

func NewReferenceSpecificatonProvider() ocmcomm.ElementSpecificationsProvider {
	a := &ReferenceResourceSpecificationProvider{
		ElementMetaDataSpecificationsProvider: ocmcomm.NewElementMetaDataSpecificationsProvider("reference", addMeta,
			flagsets.NewStringOptionType("component", "component name"),
		),
	}
	return a
}

func addMeta(opts flagsets.ConfigOptions, config flagsets.Config) error {
	flagsets.AddFieldByOption(opts, "component", config, "componentName")
	return nil
}

func (a *ReferenceResourceSpecificationProvider) Description() string {
	return a.ElementMetaDataSpecificationsProvider.Description() + `
The component name can be specified with the option <code>--component</code>. 
Therefore, basic references not requiring any additional labels or extra
identities can just be specified by those simple value options without the need
for the YAML option.
`
}

func (a *ReferenceResourceSpecificationProvider) Get() (string, error) {
	data, err := a.ParsedMeta()
	if err != nil {
		return "", err
	}

	//nolint:errchkjson // We don't care about this error.
	r, _ := json.Marshal(data)
	return string(r), nil
}

func (a *ReferenceResourceSpecificationProvider) Resources() ([]addhdlrs.ElementSource, error) {
	if !a.IsSpecified() {
		return nil, nil
	}
	return []addhdlrs.ElementSource{a}, nil
}
