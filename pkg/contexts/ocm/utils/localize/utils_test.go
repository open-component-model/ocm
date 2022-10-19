// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package localize_test

import (
	"fmt"
	"strings"

	. "github.com/onsi/gomega"

	"github.com/mandelsoft/vfs/pkg/vfs"

	"github.com/open-component-model/ocm/pkg/contexts/ocm/utils/localize"
	"github.com/open-component-model/ocm/pkg/runtime"
)

func Localizations(data string) []localize.Localization {
	var v []localize.Localization
	Expect(runtime.DefaultYAMLEncoding.Unmarshal([]byte(data), &v)).To(Succeed())
	return v
}

func Configurations(data string) []localize.Configuration {
	var v []localize.Configuration
	Expect(runtime.DefaultYAMLEncoding.Unmarshal([]byte(data), &v)).To(Succeed())
	return v
}

func Substitutions(data string) localize.Substitutions {
	var v localize.Substitutions
	Expect(runtime.DefaultYAMLEncoding.Unmarshal([]byte(data), &v)).To(Succeed())
	return v
}

func InstRules(data string) *localize.InstantiationRules {
	var v localize.InstantiationRules
	Expect(runtime.DefaultYAMLEncoding.Unmarshal([]byte(data), &v)).To(Succeed())
	return &v
}

func CheckFile(path string, fs vfs.FileSystem, content string) {
	data, err := vfs.ReadFile(fs, path)
	ExpectWithOffset(1, err).To(Succeed())
	fmt.Printf("\n%s\n", string(data))
	ExpectWithOffset(1, strings.Trim(string(data), "\n")).To(Equal(strings.Trim(content, "\n")))
}
