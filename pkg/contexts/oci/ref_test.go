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

package oci_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	oci2 "github.com/open-component-model/ocm/pkg/contexts/oci"
	"github.com/opencontainers/go-digest"
)

func CheckRef(ref string, exp *oci2.RefSpec) {
	spec, err := oci2.ParseRef(ref)
	if exp == nil {
		ExpectWithOffset(1, err).To(HaveOccurred())
	} else {
		ExpectWithOffset(1, err).To(Succeed())
		ExpectWithOffset(1, spec).To(Equal(*exp))
	}
}

func CheckRepo(ref string, exp *oci2.UniformRepositorySpec) {
	spec, err := oci2.ParseRepo(ref)
	if exp == nil {
		ExpectWithOffset(1, err).To(HaveOccurred())
	} else {
		ExpectWithOffset(1, err).To(Succeed())
		ExpectWithOffset(1, spec).To(Equal(*exp))
	}
}

var _ = Describe("ref parsing", func() {
	digest := digest.Digest("sha256:3d05e105e350edf5be64fe356f4906dd3f9bf442a279e4142db9879bba8e677a")
	tag := "v1"

	ghcr := oci2.UniformRepositorySpec{Host: "ghcr.io"}
	docker := oci2.UniformRepositorySpec{Host: "docker.io"}

	It("succeeds for repository", func() {
		CheckRef("::ghcr.io/", &oci2.RefSpec{UniformRepositorySpec: ghcr})
	})
	It("succeeds", func() {

		CheckRef("ubuntu", &oci2.RefSpec{UniformRepositorySpec: docker, Repository: "library/ubuntu"})
		CheckRef("ubuntu:v1", &oci2.RefSpec{UniformRepositorySpec: docker, Repository: "library/ubuntu", Tag: &tag})
		CheckRef("test/ubuntu", &oci2.RefSpec{UniformRepositorySpec: docker, Repository: "test/ubuntu"})
		CheckRef("test/ubuntu:v1", &oci2.RefSpec{UniformRepositorySpec: docker, Repository: "test/ubuntu", Tag: &tag})
		CheckRef("ghcr.io/test/ubuntu", &oci2.RefSpec{UniformRepositorySpec: ghcr, Repository: "test/ubuntu"})
		CheckRef("ghcr.io:8080/test/ubuntu", &oci2.RefSpec{UniformRepositorySpec: oci2.UniformRepositorySpec{Host: "ghcr.io:8080"}, Repository: "test/ubuntu"})
		CheckRef("ghcr.io/test/ubuntu:v1", &oci2.RefSpec{UniformRepositorySpec: ghcr, Repository: "test/ubuntu", Tag: &tag})
		CheckRef("ghcr.io/test/ubuntu@sha256:3d05e105e350edf5be64fe356f4906dd3f9bf442a279e4142db9879bba8e677a", &oci2.RefSpec{UniformRepositorySpec: ghcr, Repository: "test/ubuntu", Digest: &digest})
		CheckRef("ghcr.io/test/ubuntu:v1@sha256:3d05e105e350edf5be64fe356f4906dd3f9bf442a279e4142db9879bba8e677a", &oci2.RefSpec{UniformRepositorySpec: ghcr, Repository: "test/ubuntu", Tag: &tag, Digest: &digest})
		CheckRef("type::https://ghcr.io/repo/repo:v1@"+digest.String(), &oci2.RefSpec{
			UniformRepositorySpec: oci2.UniformRepositorySpec{
				Type:   "type",
				Scheme: "https",
				Host:   "ghcr.io",
				Info:   "",
			},
			Repository: "repo/repo",
			Tag:        &tag,
			Digest:     &digest,
		})
		CheckRef("directory::a/b", &oci2.RefSpec{
			UniformRepositorySpec: oci2.UniformRepositorySpec{
				Type:   "directory",
				Scheme: "",
				Host:   "",
				Info:   "a/b",
			},
			Repository: "",
		})

		CheckRef("a/b//", &oci2.RefSpec{
			UniformRepositorySpec: oci2.UniformRepositorySpec{
				Type:   "",
				Scheme: "",
				Host:   "",
				Info:   "a/b",
			},
			Repository: "",
		})

		CheckRef("directory::a/b//c/d", &oci2.RefSpec{
			UniformRepositorySpec: oci2.UniformRepositorySpec{
				Type:   "directory",
				Scheme: "",
				Host:   "",
				Info:   "a/b",
			},
			Repository: "c/d",
		})

		CheckRef("oci::ghcr.io", &oci2.RefSpec{
			UniformRepositorySpec: oci2.UniformRepositorySpec{
				Type:   "oci",
				Scheme: "",
				Host:   "ghcr.io",
				Info:   "",
			},
			Repository: "",
		})
	})

	It("fails", func() {
		CheckRef("https://ubuntu", nil)
		CheckRef("ubuntu@4711", nil)
		CheckRef("test/ubuntu@4711", nil)
		CheckRef("test/ubuntu:v1@4711", nil)
		CheckRef("ghcr.io/test/ubuntu:v1@4711", nil)

	})
	It("repo", func() {
		CheckRepo("ghcr.io", &oci2.UniformRepositorySpec{
			Host: "ghcr.io",
		})
		CheckRepo("https://ghcr.io", &oci2.UniformRepositorySpec{
			Scheme: "https",
			Host:   "ghcr.io",
		})
		CheckRepo("alias", &oci2.UniformRepositorySpec{
			Info: "alias",
		})
		CheckRepo("tar::a/b.tar", &oci2.UniformRepositorySpec{
			Type: "tar",
			Info: "a/b.tar",
		})
		CheckRepo("a/b.tar", &oci2.UniformRepositorySpec{
			Info: "a/b.tar",
		})
	})

})
