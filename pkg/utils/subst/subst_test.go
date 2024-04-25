// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package subst

import (
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

		It("handles simple block value substitution on yaml", func() {
			data := `
data:
  value1: null
  value2: orig2
`
			content, err := Parse([]byte(data))
			Expect(err).To(Succeed())

			Expect(content.SubstituteByValue("data.value1", `|
    -----BEGIN CERTIFICATE-----
    MIIDjjCCAnagAwIBAgIQAzrx5qcRqaC7KGSxHQn65TANBgkqhkiG9w0BAQsFADBh
    MQswCQYDVQQGEwJVUzEVMBMGA1UEChMMRGlnaUNlcnQgSW5jMRkwFwYDVQQLExB3
    d3cuZGlnaWNlcnQuY29tMSAwHgYDVQQDExdEaWdpQ2VydCBHbG9iYWwgUm9vdCBH
    MjAeFw0xMzA4MDExMjAwMDBaFw0zODAxMTUxMjAwMDBaMGExCzAJBgNVBAYTAlVT
    MRUwEwYDVQQKEwxEaWdpQ2VydCBJbmMxGTAXBgNVBAsTEHd3dy5kaWdpY2VydC5j
    b20xIDAeBgNVBAMTF0RpZ2lDZXJ0IEdsb2JhbCBSb290IEcyMIIBIjANBgkqhkiG
    9w0BAQEFAAOCAQ8AMIIBCgKCAQEAuzfNNNx7a8myaJCtSnX/RrohCgiN9RlUyfuI
    2/Ou8jqJkTx65qsGGmvPrC3oXgkkRLpimn7Wo6h+4FR1IAWsULecYxpsMNzaHxmx
    1x7e/dfgy5SDN67sH0NO3Xss0r0upS/kqbitOtSZpLYl6ZtrAGCSYP9PIUkY92eQ
    q2EGnI/yuum06ZIya7XzV+hdG82MHauVBJVJ8zUtluNJbd134/tJS7SsVQepj5Wz
    tCO7TG1F8PapspUwtP1MVYwnSlcUfIKdzXOS0xZKBgyMUNGPHgm+F6HmIcr9g+UQ
    vIOlCsRnKPZzFBQ9RnbDhxSJITRNrw9FDKZJobq7nMWxM4MphQIDAQABo0IwQDAP
    BgNVHRMBAf8EBTADAQH/MA4GA1UdDwEB/wQEAwIBhjAdBgNVHQ4EFgQUTiJUIBiV
    5uNu5g/6+rkS7QYXjzkwDQYJKoZIhvcNAQELBQADggEBAGBnKJRvDkhj6zHd6mcY
    1Yl9PMWLSn/pvtsrF9+wX3N3KjITOYFnQoQj8kVnNeyIv/iPsGEMNKSuIEyExtv4
    NeF22d+mQrvHRAiGfzZ0JFrabA0UWTW98kndth/Jsw1HKj2ZL7tcu7XUIOGZX1NG
    Fdtom/DzMNU+MeKNhJ7jitralj41E6Vf8PlwUHBHQRFXGU7Aj64GxJUTFy8bJZ91
    8rGOmaFvE7FBcf6IKshPECBV1/MUReXgRPTqh5Uykw7+U0b6LJ3/iyK5S9kJRaTe
    pLiaWN0bfVKfjllDiIGknibVb63dDcY3fe0Dkhvld1927jyNxF1WW6LZZm6zNTfl
    MrY=
    -----END CERTIFICATE-----`)).To(Succeed())

			result, err := content.Content()
			Expect(err).To(Succeed())

			expected := `
data:
  value1: |
    -----BEGIN CERTIFICATE-----
    MIIDjjCCAnagAwIBAgIQAzrx5qcRqaC7KGSxHQn65TANBgkqhkiG9w0BAQsFADBh
    MQswCQYDVQQGEwJVUzEVMBMGA1UEChMMRGlnaUNlcnQgSW5jMRkwFwYDVQQLExB3
    d3cuZGlnaWNlcnQuY29tMSAwHgYDVQQDExdEaWdpQ2VydCBHbG9iYWwgUm9vdCBH
    MjAeFw0xMzA4MDExMjAwMDBaFw0zODAxMTUxMjAwMDBaMGExCzAJBgNVBAYTAlVT
    MRUwEwYDVQQKEwxEaWdpQ2VydCBJbmMxGTAXBgNVBAsTEHd3dy5kaWdpY2VydC5j
    b20xIDAeBgNVBAMTF0RpZ2lDZXJ0IEdsb2JhbCBSb290IEcyMIIBIjANBgkqhkiG
    9w0BAQEFAAOCAQ8AMIIBCgKCAQEAuzfNNNx7a8myaJCtSnX/RrohCgiN9RlUyfuI
    2/Ou8jqJkTx65qsGGmvPrC3oXgkkRLpimn7Wo6h+4FR1IAWsULecYxpsMNzaHxmx
    1x7e/dfgy5SDN67sH0NO3Xss0r0upS/kqbitOtSZpLYl6ZtrAGCSYP9PIUkY92eQ
    q2EGnI/yuum06ZIya7XzV+hdG82MHauVBJVJ8zUtluNJbd134/tJS7SsVQepj5Wz
    tCO7TG1F8PapspUwtP1MVYwnSlcUfIKdzXOS0xZKBgyMUNGPHgm+F6HmIcr9g+UQ
    vIOlCsRnKPZzFBQ9RnbDhxSJITRNrw9FDKZJobq7nMWxM4MphQIDAQABo0IwQDAP
    BgNVHRMBAf8EBTADAQH/MA4GA1UdDwEB/wQEAwIBhjAdBgNVHQ4EFgQUTiJUIBiV
    5uNu5g/6+rkS7QYXjzkwDQYJKoZIhvcNAQELBQADggEBAGBnKJRvDkhj6zHd6mcY
    1Yl9PMWLSn/pvtsrF9+wX3N3KjITOYFnQoQj8kVnNeyIv/iPsGEMNKSuIEyExtv4
    NeF22d+mQrvHRAiGfzZ0JFrabA0UWTW98kndth/Jsw1HKj2ZL7tcu7XUIOGZX1NG
    Fdtom/DzMNU+MeKNhJ7jitralj41E6Vf8PlwUHBHQRFXGU7Aj64GxJUTFy8bJZ91
    8rGOmaFvE7FBcf6IKshPECBV1/MUReXgRPTqh5Uykw7+U0b6LJ3/iyK5S9kJRaTe
    pLiaWN0bfVKfjllDiIGknibVb63dDcY3fe0Dkhvld1927jyNxF1WW6LZZm6zNTfl
    MrY=
    -----END CERTIFICATE-----
  value2: orig2
`
			Expect(strings.Trim(string(result), "\n")).To(Equal(strings.Trim(expected, "\n")))
		})

		It("handles complex value substitution on yaml 1", func() {
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

		It("handles complex value substitution on yaml 2", func() {
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

		Expect(strings.Trim(string(result), "\n")).To(Equal(strings.Trim(`
data:
  value1:
    value1: v1
    value2: ""
  value2: orig2
`, "\n")))
	})
})
