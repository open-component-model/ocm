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

package v1

import (
	"regexp"
	"unicode"
)

const IdentityKeyValidationErrMsg string = "a identity label or name must consist of lower case alphanumeric characters, '-', '_' or '+', and must start and end with an alphanumeric character"

var identityKeyValidationRegexp = regexp.MustCompile("^[a-z0-9]([-_+a-z0-9]*[a-z0-9])?$")

func IsIdentity(s string) bool {
	return identityKeyValidationRegexp.Match([]byte(s))
}

// IsAscii checks whether a string only contains ascii characters.
func IsASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] > unicode.MaxASCII {
			return false
		}
	}
	return true
}
