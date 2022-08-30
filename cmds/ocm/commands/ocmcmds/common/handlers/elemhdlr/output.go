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

package elemhdlr

import (
	"encoding/json"

	"github.com/sirupsen/logrus"

	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/runtime"
)

var MetaOutput = []string{"NAME", "VERSION", "IDENTITY"}

func MapMetaOutput(e interface{}) []string {
	p := e.(*Object)
	m := p.Element.GetMeta()
	id := p.Id.Copy()
	id.Remove(metav1.SystemIdentityName)
	return []string{m.Name, m.Version, id.String()}
}

var AccessOutput = []string{"ACCESSTYPE", "ACCESSSPEC"}

func MapAccessOutput(e compdesc.AccessSpec) []string {
	a := ""
	data, err := json.Marshal(e)
	if err != nil {
		a = "invalid: " + err.Error()
	} else {
		var un map[string]interface{}
		err := json.Unmarshal(data, &un)
		if err != nil {
			a = "invalid: " + err.Error()
		} else {
			delete(un, runtime.ATTR_TYPE)

			data, err = json.Marshal(un)
			if err != nil {
				logrus.Error(err)
			}

			a = string(data)
		}
	}
	return []string{e.GetKind(), a}
}
