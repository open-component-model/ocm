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

package localize_test

import (
	"github.com/ghodss/yaml"
	"github.com/mandelsoft/vfs/pkg/vfs"
	. "github.com/onsi/gomega"

	"github.com/open-component-model/ocm/pkg/contexts/ocm/utils/localize"
)

func Localizations(data string) []localize.Localization {
	var v []localize.Localization
	Expect(yaml.Unmarshal([]byte(data), &v)).To(Succeed())
	return v
}

func Configurations(data string) []localize.Configuration {
	var v []localize.Configuration
	Expect(yaml.Unmarshal([]byte(data), &v)).To(Succeed())
	return v
}

func Substitutions(data string) localize.Substitutions {
	var v localize.Substitutions
	Expect(yaml.Unmarshal([]byte(data), &v)).To(Succeed())
	return v
}

func InstRules(data string) *localize.InstantiationRules {
	var v localize.InstantiationRules
	Expect(yaml.Unmarshal([]byte(data), &v)).To(Succeed())
	return &v
}

func CheckFile(path string, fs vfs.FileSystem, content string) {
	data, err := vfs.ReadFile(fs, path)
	ExpectWithOffset(1, err).To(Succeed())
	ExpectWithOffset(1, "\n"+string(data)).To(Equal(content))
}
