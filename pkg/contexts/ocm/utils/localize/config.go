// Copyright 2022 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package localize

import (
	"fmt"

	"github.com/mandelsoft/spiff/spiffing"

	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/utils"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/runtime"
	"github.com/open-component-model/ocm/pkg/spiff"
)

func Configure(mappings []Configuration, cursubst []Substitution, cv ocm.ComponentVersionAccess, resolver ocm.ComponentVersionResolver, template []byte, config []byte, libraries []metav1.ResourceReference, schemedata []byte) (Substitutions, error) {
	var err error

	if len(mappings) == 0 {
		return nil, nil
	}
	if len(config) == 0 {
		if len(schemedata) > 0 {
			err = spiff.ValidateByScheme([]byte("{}"), schemedata)
			if err != nil {
				return nil, errors.Wrapf(err, "config validation failed")
			}
		}
		if len(template) == 0 {
			return nil, nil
		}
	}

	stubs := spiff.Options{}
	for i, lib := range libraries {
		res, eff, err := utils.ResolveResourceReference(cv, lib, resolver)
		if err != nil {
			return nil, errors.ErrNotFound("library resource %s not found", lib.String())
		}
		defer eff.Close()
		m, err := res.AccessMethod()
		if err != nil {
			return nil, errors.ErrNotFound("cannot access library resource", lib.String())
		}
		data, err := m.Get()
		m.Close()
		if err != nil {
			return nil, errors.ErrNotFound("cannot access library resource", lib.String())
		}
		stubs.Add(spiff.StubData(fmt.Sprintf("spiff lib%d", i), data))
	}

	if len(schemedata) > 0 {
		err = spiff.ValidateByScheme(config, schemedata)
		if err != nil {
			return nil, errors.Wrapf(err, "validation failed")
		}
	}

	list := []interface{}{}
	for _, e := range cursubst {
		// TODO: escape spiff expressions, but should not occur, so omit it so far
		list = append(list, e)
	}
	for _, e := range mappings {
		list = append(list, e)
	}

	var temp map[string]interface{}
	if len(template) == 0 {
		temp = map[string]interface{}{
			"adjustments": list,
		}
	} else {
		if err = runtime.DefaultYAMLEncoding.Unmarshal(template, &temp); err != nil {
			return nil, errors.Wrapf(err, "cannot unmarshal template")
		}
		if _, ok := temp["adjustments"]; ok {
			return nil, errors.Newf("template may not contain 'adjustments'")
		}
		temp["adjustments"] = list
	}

	if _, ok := temp["utilities"]; !ok {
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
	}
	err = runtime.DefaultYAMLEncoding.Unmarshal(config, &subst)
	return subst.Adjustments, err
}
