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

package subst

import (
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type complex struct {
	Value1 string `json:"value1"`
	Value2 string `json:"value2"`
}

var _ = Describe("value substitution", func() {
	Context("by values", func() {
		It("handles simple value substitution on yaml", func() {
			data := `
data:
  value1: orig1
  value2: orig2
`
			content, err := Parse([]byte(data))
			Expect(err).To(Succeed())

			Expect(content.SubstituteByValue("data.value1", "v1")).To(Succeed())

			result, err := content.Content()
			Expect(err).To(Succeed())

			Expect(strings.Trim(string(result), "\n")).To(Equal(strings.Trim(`
data:
  value1: v1
  value2: orig2
`, "\n")))
		})

		It("handles complex value substitution on yaml", func() {
			data := `
data:
  value1: orig1
  value2: orig2
`
			content, err := Parse([]byte(data))
			Expect(err).To(Succeed())

			Expect(content.SubstituteByValue("data.value1", &complex{Value1: "v1"})).To(Succeed())

			result, err := content.Content()
			Expect(err).To(Succeed())

			fmt.Printf("\n%s\n", string(result))
			Expect(strings.Trim(string(result), "\n")).To(Equal(strings.Trim(`
data:
  value1:
    value1: v1
    value2: ""
  value2: orig2
`, "\n")))
		})
	})

	Context("by data", func() {
		It("handles simple value substitution on yaml", func() {
			data := `
data:
  value1: orig1
  value2: orig2
`
			content, err := Parse([]byte(data))
			Expect(err).To(Succeed())

			Expect(content.SubstituteByData("data.value1", []byte("\"v1\""))).To(Succeed())

			result, err := content.Content()
			Expect(err).To(Succeed())

			Expect(strings.Trim(string(result), "\n")).To(Equal(strings.Trim(`
data:
  value1: v1
  value2: orig2
`, "\n")))
		})

		It("handles complex value substitution on yaml", func() {
			value := `
value1: v1
value2: ""
`
			data := `
data:
  value1: orig1
  value2: orig2
`
			content, err := Parse([]byte(data))
			Expect(err).To(Succeed())

			Expect(content.SubstituteByData("data.value1", []byte(value))).To(Succeed())

			result, err := content.Content()
			Expect(err).To(Succeed())

			fmt.Printf("\n%s\n", string(result))
			Expect(strings.Trim(string(result), "\n")).To(Equal(strings.Trim(`
data:
  value1:
    value1: v1
    value2: ""
  value2: orig2
`, "\n")))
		})
	})

	It("handles complex value substitution on json", func() {
		value := `
value1: v1
value2: ""
`
		data := `
{ "data": {
    "value1": "orig1",
    "value2": "orig2"
  }
}
`
		content, err := Parse([]byte(data))
		Expect(err).To(Succeed())

		Expect(content.SubstituteByData("data.value1", []byte(value))).To(Succeed())

		result, err := content.Content()
		Expect(err).To(Succeed())

		fmt.Printf("\n%s\n", string(result))
		Expect(strings.Trim(string(result), "\n")).To(Equal(strings.Trim(`
{"data": {"value1": {"value1": "v1", "value2": ""}, "value2": "orig2"}}
`, "\n")))
	})

	/*
			It("handles json/yaml mix", func() {
				value := `
		value1: v1
		value2: ""
		`
				data := `
		data: {
		    "value1": "orig1",
		    "value2": "orig2"
		}
		`
				content, err := Parse([]byte(data))
				Expect(err).To(Succeed())

				Expect(content.SubstituteByData("data.value1", []byte(value))).To(Succeed())

				result, err := content.Content()
				Expect(err).To(Succeed())

				fmt.Printf("\n%s\n", string(result))
				Expect(strings.Trim(string(result), "\n")).To(Equal(strings.Trim(`
		data: {"value1":
		        value1: v1
		        value2: "", "value2": "orig2"}
		`, "\n")))

				// THIS IS COMPLETE BULLSHIT, it is no yaml
				var m map[string]interface{}
				err = runtime.DefaultYAMLEncoding.Unmarshal(result, &m)
				fmt.Printf("%s\n", err)
				Expect(err).To(HaveOccurred())

				err = yaml.Unmarshal(result, &m)
				Expect(err).To(HaveOccurred())
			})
	*/
	It("handles json/yaml mix", func() {
		value := `
value1: v1
value2: ""
`
		data := `
data: {
    "value1": "orig1",
    "value2": "orig2"
}
`
		content, err := Parse([]byte(data))
		Expect(err).To(Succeed())

		Expect(content.SubstituteByData("data.value1", []byte(value))).To(Succeed())

		result, err := content.Content()
		Expect(err).To(Succeed())

		fmt.Printf("\n%s\n", string(result))
		Expect(strings.Trim(string(result), "\n")).To(Equal(strings.Trim(`
data:
  value1:
    value1: v1
    value2: ""
  value2: orig2
`, "\n")))
	})
})
