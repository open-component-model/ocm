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

package cobrautils

import (
	"strings"

	"golang.org/x/text/cases"
)

var templatefuncs = map[string]interface{}{
	"indent":      indent,
	"skipCommand": skipCommand,
	"soleCommand": soleCommand,
	"title":       cases.Title,
}

func soleCommand(s string) bool {
	return !strings.Contains(s, " ")
}

func skipCommand(s string) string {
	i := strings.Index(s, " ")
	if i < 0 {
		return ""
	}
	for ; i < len(s); i++ {
		if s[i] != ' ' {
			return s[i:]
		}
	}
	return ""
}

func indent(n int, s string) string {
	gap := ""
	for ; n > 0; n-- {
		gap += " "
	}
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		if len(l) > 0 {
			lines[i] = gap + lines[i]
		}
	}
	return strings.Join(lines, "\n")
}
