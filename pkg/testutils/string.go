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

package testutils

import (
	"fmt"
	"strings"

	"github.com/onsi/gomega/format"
	"github.com/onsi/gomega/types"
)

// StringEqualTrimmedWithContext compares two trimmed strings and provides the complete actual value
// as error context.
// It is an error for actual to be nil.  Use BeNil() instead.
func StringEqualTrimmedWithContext(expected string) types.GomegaMatcher {
	return &StringEqualMatcher{
		Expected: expected,
		Trim:     true,
	}
}

// StringEqualWithContext compares two strings and provides the complete actual value
// as error context.
// It is an error for actual to be nil.  Use BeNil() instead.
func StringEqualWithContext(expected string) types.GomegaMatcher {
	return &StringEqualMatcher{
		Expected: expected,
	}
}

type StringEqualMatcher struct {
	Expected string
	Trim     bool
}

func (matcher *StringEqualMatcher) Match(actual interface{}) (success bool, err error) {
	if actual == nil {
		return false, fmt.Errorf("Refusing to compare <nil> to <string>.")
	}

	s, ok := actual.(string)
	if !ok {
		return false, fmt.Errorf("Actual value is no string, but a %T.", actual)
	}
	if matcher.Trim {
		return strings.TrimSpace(s) == strings.TrimSpace(matcher.Expected), nil
	}
	return s == matcher.Expected, nil
}

func (matcher *StringEqualMatcher) FailureMessage(actual interface{}) (message string) {
	actualString, actualOK := actual.(string)
	if actualOK {
		if matcher.Trim {
			return "Found\n" + actualString + "\n" + format.MessageWithDiff(strings.TrimSpace(actualString), "to equal", strings.TrimSpace(matcher.Expected))
		}
		return "Found\n" + actualString + "\n" + format.MessageWithDiff(actualString, "to equal", matcher.Expected)
	}

	return format.Message(actual, "to equal", matcher.Expected)
}

func (matcher *StringEqualMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return format.Message(actual, "not to equal", matcher.Expected)
}
