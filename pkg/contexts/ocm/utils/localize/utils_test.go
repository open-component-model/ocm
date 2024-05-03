// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package localize_test

import (
	. "github.com/onsi/gomega"

	"github.com/mandelsoft/vfs/pkg/vfs"

	"github.com/open-component-model/ocm/pkg/contexts/ocm/utils/localize"
	"github.com/open-component-model/ocm/pkg/runtime"
)

func UnmarshalLocalizations(data string) []localize.Localization {
	var v []localize.Localization
	Expect(runtime.DefaultYAMLEncoding.Unmarshal([]byte(data), &v)).To(Succeed())
	return v
}

func UnmarshalConfigurations(data string) []localize.Configuration {
	var v []localize.Configuration
	Expect(runtime.DefaultYAMLEncoding.Unmarshal([]byte(data), &v)).To(Succeed())
	return v
}

func UnmarshalSubstitutions(data string) localize.Substitutions {
	var v localize.Substitutions
	Expect(runtime.DefaultYAMLEncoding.Unmarshal([]byte(data), &v)).To(Succeed())
	return v
}

func UnmarshalImageMappings(data string) localize.ImageMappings {
	var v localize.ImageMappings
	Expect(runtime.DefaultYAMLEncoding.Unmarshal([]byte(data), &v)).To(Succeed())
	return v
}

func UnmarshalValueMappings(data string) localize.ValueMappings {
	var v localize.ValueMappings
	Expect(runtime.DefaultYAMLEncoding.Unmarshal([]byte(data), &v)).To(Succeed())
	return v
}

func UnmarshalInstRules(data string) *localize.InstantiationRules {
	var v localize.InstantiationRules
	Expect(runtime.DefaultYAMLEncoding.Unmarshal([]byte(data), &v)).To(Succeed())
	return &v
}

func CheckYAMLFile(path string, fs vfs.FileSystem, content string) {
	data, err := vfs.ReadFile(fs, path)
	ExpectWithOffset(1, err).To(Succeed())
	ExpectWithOffset(1, string(data)).To(MatchYAML(content))
}

func CheckJSONFile(path string, fs vfs.FileSystem, content string) {
	data, err := vfs.ReadFile(fs, path)
	ExpectWithOffset(1, err).To(Succeed())
	ExpectWithOffset(1, string(data)).To(MatchJSON(content))
}
