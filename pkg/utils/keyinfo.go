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

package utils

import (
	"fmt"
)

type KeyInfo interface {
	GetKey() string
	GetDescription() string
}

func FormatList(def string, elems ...KeyInfo) string {
	names := ""
	for _, n := range elems {
		add := ""
		if n.GetKey() == def {
			add = " (default)"
		}
		names = fmt.Sprintf("%s\n  - <code>%s</code>:%s %s", names, n.GetKey(), add, n.GetDescription())
	}
	return names
}
