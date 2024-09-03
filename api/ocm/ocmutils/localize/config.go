package localize

import (
	"fmt"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/sliceutils"
	"github.com/mandelsoft/spiff/spiffing"

	"ocm.software/ocm/api/ocm"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/resourcerefs"
	"ocm.software/ocm/api/utils/runtime"
	"ocm.software/ocm/api/utils/spiff"
)

func Configure(
	mappings []Configuration, cursubst []Substitution,
	cv ocm.ComponentVersionAccess, resolver ocm.ComponentVersionResolver,
	template []byte, config []byte, libraries []metav1.ResourceReference, schemedata []byte,
) (Substitutions, error) {
	var err error

	if len(mappings) == 0 {
		return nil, nil
	}
	if len(config) == 0 {
		if len(schemedata) > 0 {
			if err = spiff.ValidateByScheme([]byte("{}"), schemedata); err != nil {
				return nil, errors.Wrapf(err, "config validation failed")
			}
		}
		if len(template) == 0 {
			return nil, nil
		}
	}

	stubs := spiff.Options{}
	for i, lib := range libraries {
		opt, err := func() (spiff.OptionFunction, error) {
			res, eff, err := resourcerefs.ResolveResourceReference(cv, lib, resolver)
			if err != nil {
				return nil, errors.ErrNotFound("library resource %s not found", lib.String())
			}
			defer eff.Close()
			m, err := res.AccessMethod()
			if err != nil {
				return nil, errors.ErrNotFound("error accessing access method for library resource", lib.String())
			}
			data, err := m.Get()
			m.Close()
			if err != nil {
				return nil, errors.ErrNotFound("cannot access library resource", lib.String())
			}
			return spiff.StubData(fmt.Sprintf("spiff lib%d", i), data), nil
		}()
		if err != nil {
			return nil, err
		}
		stubs.Add(opt)
	}

	if len(schemedata) > 0 {
		err = spiff.ValidateByScheme(config, schemedata)
		if err != nil {
			return nil, errors.Wrapf(err, "validation failed")
		}
	}

	extlist := []interface{}{}
	for _, e := range cursubst {
		// TODO: escape spiff expressions, but should not occur, so omit it so far
		extlist = append(extlist, e)
	}

	cfglist := []interface{}{}
	for _, e := range mappings {
		cfglist = append(cfglist, e)
	}

	var temp map[string]interface{}
	if len(template) == 0 {
		temp = map[string]interface{}{
			"adjustments": extlist,
			"configRules": cfglist,
		}
	} else {
		if err := runtime.DefaultYAMLEncoding.Unmarshal(template, &temp); err != nil {
			return nil, errors.Wrapf(err, "cannot unmarshal template")
		}

		if _, ok := temp["adjustments"]; ok {
			return nil, errors.Newf("template may not contain 'adjustments'")
		}
		temp["adjustments"] = extlist

		if cur, ok := temp["configRules"]; ok {
			if l, ok := cur.([]interface{}); !ok {
				return nil, errors.Newf("node 'configRules' in template must be a list of configuration requests")
			} else {
				temp["configRules"] = sliceutils.CopyAppend(l, cfglist...)
			}
		} else {
			temp["configRules"] = cfglist
		}
	}

	if _, ok := temp["utilities"]; !ok {
		// prepare merging of spiff libraries using the node utilities as root path
		temp["utilities"] = ""
	}

	template, err = runtime.DefaultJSONEncoding.Marshal(temp)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot marshal adjustments")
	}

	config, err = spiff.CascadeWith(spiff.TemplateData("adjustments", template), stubs, spiff.Values(config), spiff.Mode(spiffing.MODE_PRIVATE))
	if err != nil {
		return nil, errors.Wrapf(err, "error processing template")
	}

	var subst struct {
		Adjustments Substitutions `json:"adjustments,omitempty"`
		ConfigRules Substitutions `json:"configRules,omitempty"`
	}
	err = runtime.DefaultYAMLEncoding.Unmarshal(config, &subst)
	return sliceutils.CopyAppend(subst.Adjustments, subst.ConfigRules...), err
}
