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
	"regexp"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/open-component-model/ocm/pkg/runtime"
)

func TestCompName(dataBytes []byte, err error) {
	ExpectWithOffset(1, err).To(Succeed())

	Context("component name validation", func() {
		var scheme map[string]interface{}
		Expect(runtime.DefaultYAMLEncoding.Unmarshal(dataBytes, &scheme)).To(Succeed())

		pattern := scheme["definitions"].(map[string]interface{})["componentName"].(map[string]interface{})["pattern"].(string)

		fmt.Printf("pattern=%s\n", pattern)

		expr, err := regexp.Compile(pattern)
		Expect(err).To(Succeed())

		Check := func(s string, exp bool) {
			if expr.MatchString(s) != exp {
				Fail(fmt.Sprintf("%s[%t] failed\n", s, exp), 1)
			}
		}

		It("parsed valid names", func() {
			Check("github.wdf.sap.corp/kubernetes/landscape-setup", true)
			Check("weave.works/registry/app", true)
			Check("internal.github.org/registry/app", true)
			Check("a.de/c", true)
			Check("a.de/c/d/e-f", true)
			Check("a.de/c/d/e_f", true)
			Check("a.de/c/d/e", true)
			Check("a.de/c/d/e.f", true)
		})

		It("rejects invalid names", func() {
			Check("a.de/", false)
			Check("a.de/a/", false)
			Check("a.de//a", false)
			Check("a.de/a.", false)
		})
	})
}
