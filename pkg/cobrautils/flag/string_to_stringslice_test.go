// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package flag

import (
	"fmt"
	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("", func() {
	FIt("test", func() {
		s := stringToStringSliceValue[map[string][]string]{}
		err := s.Set("key1=val1")
		fmt.Println(err)
		//fmt.Println(s.value)
	})
})
