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

package grammar

import (
	"regexp"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	tool "github.com/open-component-model/ocm/pkg/regex"
)

func TestConfig(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "OCI Test Suite")
}

func CheckURI(ref string, parts ...string) {
	Check(ref, TypedURIRegexp, parts...)
}

func Check(ref string, exp *regexp.Regexp, parts ...string) {
	spec := exp.FindSubmatch([]byte(ref))
	if len(parts) == 0 {
		Expect(spec).To(BeNil())
	} else {
		result := make([]string, len(spec))
		for i, v := range spec {
			result[i] = string(v)
		}
		Expect(result).To(Equal(append([]string{ref}, parts...)))
	}
}

func Type(t string) string {
	if t == "" {
		return t
	}
	return t + "::"
}
func Sub(t string) string {
	if t == "" {
		return t
	}
	return "/" + t
}
func Vers(t string) string {
	if t == "" {
		return t
	}
	return ":" + t
}

var _ = Describe("ref matching", func() {

	Context("parts", func() {
		It("path port", func() {
			Check("/some/path/docker.sock:100", tool.Capture(PathPortRegexp), "/some/path/docker.sock:100")
		})
	})

	Context("types refs", func() {
		t := "DockerDaemon"
		s := "unix"
		p := "/some/path/docker.sock:100"
		r := "repo"
		v := "test"

		It("fails", func() {
			CheckURI("DockerDaemon::unix:///some/path/docker.sock:100//repo:test", t, s, p, r, v, "")
			CheckURI("DockerDaemon::unix:///some/path/docker.sock:100", t, s, p, "", "", "")
			CheckURI("DockerDaemon::unix://some/path/docker.sock:100//repo:test", t, s, p[1:], r, v, "")
		})
	})

})
