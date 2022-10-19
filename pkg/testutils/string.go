// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

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
			actualString = strings.TrimSpace(actualString)
		}
		return "Found\n" + actualString + "\n" + format.MessageWithDiff(actualString, "to equal", matcher.Expected)
	}

	return format.Message(actual, "to equal", matcher.Expected)
}

func (matcher *StringEqualMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return format.Message(actual, "not to equal", matcher.Expected)
}
