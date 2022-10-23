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

package common

import (
	"strings"

	"github.com/open-component-model/ocm/pkg/errors"
)

func ParseSettings(args []string, kinds ...string) (map[string]string, error) {
	kind := "setting"
	if len(kinds) > 0 {
		kind = kinds[0]
	}

	settings := map[string]string{}
	for _, arg := range args {
		if i := strings.Index(arg, "="); i > 0 {
			value := arg[i+1:]
			name := strings.TrimSpace(arg[0:i])
			settings[name] = value
		} else {
			return nil, errors.Newf("invalid %s %q (assignment required)", kind, arg)
		}
	}
	return settings, nil
}

func FilterSettings(args ...string) (attrs map[string]string, addArgs []string) {
	for _, arg := range args {
		if i := strings.Index(arg, "="); i > 0 {
			if attrs == nil {
				attrs = map[string]string{}
			}
			value := arg[i+1:]
			name := strings.TrimSpace(arg[0:i])
			attrs[name] = value
			continue
		}
		addArgs = append(addArgs, arg)
	}
	return attrs, addArgs
}
