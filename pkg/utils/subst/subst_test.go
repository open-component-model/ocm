// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package subst

import (
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

			Expect(string(result)).To(MatchYAML(`
data:
  value1: v1
  value2: orig2
`))
		})

		It("handles simple block value substitution on yaml", func() {
			data := `
data:
  value1: null
  value2: orig2
`
			content, err := Parse([]byte(data))
			Expect(err).To(Succeed())

			Expect(content.SubstituteByValue("data.value1", `-----BEGIN CERTIFICATE-----
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
  value1: |-
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
			Expect(string(result)).To(MatchYAML(expected))
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

			Expect(string(result)).To(MatchYAML(`
data:
  value1:
    value1: v1
    value2: ""
  value2: orig2
`))
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

			Expect(string(result)).To(MatchYAML(`
data:
  value1: v1
  value2: orig2
`))
		})

		It("open-component-model/ocm-project issue 179 store object produces invalid yaml", func() {
			value := `certificate_authority_url: https://example1.com/v1/pki/root/ca/pem
deployment: deveaws
deployment_size: xsmall
domain: example2.com
landscape_region: eu12
org: deveaws
service_hostname_suffix: .example3.com
service_kubernetes_hostname_suffix: .example4.com
space: sac`
			data := `dmi:
  gcp_project_id: unset
  orca_env_stable_values: {}
  protect_persisted_data: ""`
			content, err := Parse([]byte(data))
			Expect(err).To(Succeed())

			Expect(content.SubstituteByData("dmi.orca_env_stable_values", []byte(value))).To(Succeed())

			result, err := content.Content()
			Expect(err).To(Succeed())

			expected := `dmi:
  gcp_project_id: unset
  orca_env_stable_values:
    certificate_authority_url: https://example1.com/v1/pki/root/ca/pem
    deployment: deveaws
    deployment_size: xsmall
    domain: example2.com
    landscape_region: eu12
    org: deveaws
    service_hostname_suffix: .example3.com
    service_kubernetes_hostname_suffix: .example4.com
    space: sac
  protect_persisted_data: ""`
			Expect(string(result)).To(MatchYAML(expected))
		})

		It("Converts json subtitution to yaml when destination is yaml doc", func() {
			value := `{
  "certificate_authority_url": "https://example1.com/v1/pki/root/ca/pem",
  "deployment": "deveaws",
  "deployment_size": "xsmall",
  "domain": "example2.com",
  "landscape_region": "eu12",
  "org": "deveaws",
  "service_hostname_suffix": ".example3.com",
  "service_kubernetes_hostname_suffix": ".example4.com",
  "space": "sac"
}`
			data := `dmi:
  gcp_project_id: unset
  orca_env_stable_values: {}
  protect_persisted_data: ""`
			content, err := Parse([]byte(data))
			Expect(err).To(Succeed())

			Expect(content.SubstituteByData("dmi.orca_env_stable_values", []byte(value))).To(Succeed())

			result, err := content.Content()
			Expect(err).To(Succeed())

			expected := `dmi:
  gcp_project_id: unset
  orca_env_stable_values:
    certificate_authority_url: https://example1.com/v1/pki/root/ca/pem
    deployment: deveaws
    deployment_size: xsmall
    domain: example2.com
    landscape_region: eu12
    org: deveaws
    service_hostname_suffix: .example3.com
    service_kubernetes_hostname_suffix: .example4.com
    space: sac
  protect_persisted_data: ""`
			Expect(string(result)).To(MatchYAML(expected))

			Expect(string(result)).To(Not(ContainSubstring("{")))
		})

		It("Store sequence in yaml", func() {
			value := `- https://example1.com/v1/pki/root/ca/pem
- deveaws
- xsmall
- example2.com
- eu12
- deveaws
- .example3.com
- .example4.com
- sac`
			data := `dmi:
  gcp_project_id: unset
  orca_env_stable_values: []
  protect_persisted_data: ""`
			content, err := Parse([]byte(data))
			Expect(err).To(Succeed())

			Expect(content.SubstituteByData("dmi.orca_env_stable_values", []byte(value))).To(Succeed())

			result, err := content.Content()
			Expect(err).To(Succeed())

			expected := `dmi:
  gcp_project_id: unset
  orca_env_stable_values:
    - https://example1.com/v1/pki/root/ca/pem
    - deveaws
    - xsmall
    - example2.com
    - eu12
    - deveaws
    - .example3.com
    - .example4.com
    - sac
  protect_persisted_data: ""`
			Expect(string(result)).To(MatchYAML(expected))
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

			Expect(string(result)).To(MatchYAML(`
data:
  value1:
    value1: v1
    value2: ""
  value2: orig2
`))
		})
	})

	It("handles differing string styles", func() {
		value := `folded: > 
  foo
  bar
folded_strip: >- 
  foo
  bar
folded_keep: >+
  foo
  bar
literal: | 
  foo
  bar
literal_strip: |-
  foo
  bar
literal_keep: |+
  foo
  bar
double: "foo\nbar"
single: 'foo\nbar'`
		data := `data:
  value1: origs1
  value2: orig2`
		content, err := Parse([]byte(data))
		Expect(err).To(Succeed())

		Expect(content.SubstituteByData("data.value1", []byte(value))).To(Succeed())

		result, err := content.Content()
		Expect(err).To(Succeed())

		Expect(string(result)).To(MatchYAML(`data:
  value1:
    folded: > 
      foo
      bar
    folded_strip: >- 
      foo
      bar
    folded_keep: >+
      foo
      bar
    literal: | 
      foo
      bar
    literal_strip: |-
      foo
      bar
    literal_keep: |+
      foo
      bar
    double: "foo\nbar"
    single: 'foo\nbar'
  value2: orig2`))
	})

	It("handles non-string scalar", func() {
		value := `2`

		data := `data:
  value1: orig1
  value2: orig2`

		content, err := Parse([]byte(data))
		Expect(err).To(Succeed())

		Expect(content.SubstituteByData("data.value1", []byte(value))).To(Succeed())

		result, err := content.Content()
		Expect(err).To(Succeed())
		expected := `data:
  value1: 2
  value2: orig2`

		Expect(string(result)).To(MatchYAML(expected))
	})

	It("handles multiple updates", func() {
		value1 := `2`
		value2 := `3`

		data := `data:
  value1: orig1
  value2: orig2`

		content, err := Parse([]byte(data))
		Expect(err).To(Succeed())

		Expect(content.SubstituteByData("data.value1", []byte(value1))).To(Succeed())

		result1, err := content.Content()
		Expect(err).To(Succeed())
		expected1 := `data:
  value1: 2
  value2: orig2`

		Expect(string(result1)).To(MatchYAML(expected1))

		Expect(content.SubstituteByData("data.value2", []byte(value2))).To(Succeed())

		result2, err := content.Content()
		Expect(err).To(Succeed())

		expected2 := `data:
  value1: 2
  value2: 3`

		Expect(string(result2)).To(MatchYAML(expected2))
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
		expected := `
{"data": {"value1": {"value1": "v1", "value2": ""}, "value2": "orig2"}}
`
		Expect(string(result)).To(MatchJSON(expected))
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

		expected := `
data:
  value1:
    value1: v1
    value2: ""
  value2: orig2
`
		Expect(string(result)).To(MatchYAML(expected))
	})
})
