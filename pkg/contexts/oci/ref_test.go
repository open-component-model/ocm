// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package oci_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/opencontainers/go-digest"

	"github.com/open-component-model/ocm/v2/pkg/contexts/oci"
)

func CheckRef(ref string, exp *oci.RefSpec) {
	spec, err := oci.ParseRef(ref)
	if exp == nil {
		ExpectWithOffset(1, err).To(HaveOccurred())
	} else {
		ExpectWithOffset(1, err).To(Succeed())
		ExpectWithOffset(1, spec).To(Equal(*exp))
	}
}

func CheckRepo(ref string, exp *oci.UniformRepositorySpec) {
	spec, err := oci.ParseRepo(ref)
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

	ghcr := oci.UniformRepositorySpec{Host: "ghcr.io"}
	docker := oci.UniformRepositorySpec{Host: "docker.io"}

	It("succeeds for repository", func() {
		CheckRef("::ghcr.io/", &oci.RefSpec{UniformRepositorySpec: ghcr})
	})
	It("succeeds", func() {

		CheckRef("ubuntu", &oci.RefSpec{UniformRepositorySpec: docker, ArtSpec: oci.ArtSpec{Repository: "library/ubuntu"}})
		CheckRef("ubuntu:v1", &oci.RefSpec{UniformRepositorySpec: docker, ArtSpec: oci.ArtSpec{Repository: "library/ubuntu", Tag: &tag}})
		CheckRef("test/ubuntu", &oci.RefSpec{UniformRepositorySpec: docker, ArtSpec: oci.ArtSpec{Repository: "test/ubuntu"}})
		CheckRef("test_test/ubuntu", &oci.RefSpec{UniformRepositorySpec: docker, ArtSpec: oci.ArtSpec{Repository: "test_test/ubuntu"}})
		CheckRef("test__test/ubuntu", &oci.RefSpec{UniformRepositorySpec: docker, ArtSpec: oci.ArtSpec{Repository: "test__test/ubuntu"}})
		CheckRef("test-test/ubuntu", &oci.RefSpec{UniformRepositorySpec: docker, ArtSpec: oci.ArtSpec{Repository: "test-test/ubuntu"}})
		CheckRef("test--test/ubuntu", &oci.RefSpec{UniformRepositorySpec: docker, ArtSpec: oci.ArtSpec{Repository: "test--test/ubuntu"}})
		CheckRef("test-----test/ubuntu", &oci.RefSpec{UniformRepositorySpec: docker, ArtSpec: oci.ArtSpec{Repository: "test-----test/ubuntu"}})
		CheckRef("test/ubuntu:v1", &oci.RefSpec{UniformRepositorySpec: docker, ArtSpec: oci.ArtSpec{Repository: "test/ubuntu", Tag: &tag}})
		CheckRef("ghcr.io/test/ubuntu", &oci.RefSpec{UniformRepositorySpec: ghcr, ArtSpec: oci.ArtSpec{Repository: "test/ubuntu"}})
		CheckRef("ghcr.io/test", &oci.RefSpec{UniformRepositorySpec: ghcr, ArtSpec: oci.ArtSpec{Repository: "test"}})
		CheckRef("ghcr.io:8080/test/ubuntu", &oci.RefSpec{UniformRepositorySpec: oci.UniformRepositorySpec{Host: "ghcr.io:8080"}, ArtSpec: oci.ArtSpec{Repository: "test/ubuntu"}})
		CheckRef("ghcr.io/test/ubuntu:v1", &oci.RefSpec{UniformRepositorySpec: ghcr, ArtSpec: oci.ArtSpec{Repository: "test/ubuntu", Tag: &tag}})
		CheckRef("ghcr.io/test/ubuntu@sha256:3d05e105e350edf5be64fe356f4906dd3f9bf442a279e4142db9879bba8e677a", &oci.RefSpec{UniformRepositorySpec: ghcr, ArtSpec: oci.ArtSpec{Repository: "test/ubuntu", Digest: &digest}})
		CheckRef("ghcr.io/test/ubuntu:v1@sha256:3d05e105e350edf5be64fe356f4906dd3f9bf442a279e4142db9879bba8e677a", &oci.RefSpec{UniformRepositorySpec: ghcr, ArtSpec: oci.ArtSpec{Repository: "test/ubuntu", Tag: &tag, Digest: &digest}})
		CheckRef("test___test/ubuntu", &oci.RefSpec{
			UniformRepositorySpec: oci.UniformRepositorySpec{
				Info: "test___test/ubuntu",
			},
		})
		CheckRef("type::https://ghcr.io/repo/repo:v1@"+digest.String(), &oci.RefSpec{
			UniformRepositorySpec: oci.UniformRepositorySpec{
				Type:   "type",
				Scheme: "https",
				Host:   "ghcr.io",
				Info:   "",
			},
			ArtSpec: oci.ArtSpec{
				Repository: "repo/repo",
				Tag:        &tag,
				Digest:     &digest,
			},
		})
		CheckRef("http://127.0.0.1:443/repo/repo:v1@"+digest.String(), &oci.RefSpec{
			UniformRepositorySpec: oci.UniformRepositorySpec{
				Type:   "",
				Scheme: "http",
				Host:   "127.0.0.1:443",
				Info:   "",
			},
			ArtSpec: oci.ArtSpec{
				Repository: "repo/repo",
				Tag:        &tag,
				Digest:     &digest,
			},
		})
		CheckRef("directory::a/b", &oci.RefSpec{
			UniformRepositorySpec: oci.UniformRepositorySpec{
				Type:   "directory",
				Scheme: "",
				Host:   "",
				Info:   "a/b",
			},
			ArtSpec: oci.ArtSpec{
				Repository: "",
			},
		})
		CheckRef("ctf+directory::a/b", &oci.RefSpec{
			UniformRepositorySpec: oci.UniformRepositorySpec{
				Type:   "ctf+directory",
				Scheme: "",
				Host:   "",
				Info:   "a/b",
			},
			ArtSpec: oci.ArtSpec{
				Repository: "",
			},
		})
		CheckRef("+ctf+directory::a/b", &oci.RefSpec{
			UniformRepositorySpec: oci.UniformRepositorySpec{
				Type:            "ctf+directory",
				Scheme:          "",
				Host:            "",
				Info:            "a/b",
				CreateIfMissing: true,
			},
			ArtSpec: oci.ArtSpec{
				Repository: "",
			},
		})

		CheckRef("a/b//", &oci.RefSpec{
			UniformRepositorySpec: oci.UniformRepositorySpec{
				Type:   "",
				Scheme: "",
				Host:   "",
				Info:   "a/b",
			},
			ArtSpec: oci.ArtSpec{
				Repository: "",
			},
		})

		CheckRef("directory::a/b//c/d", &oci.RefSpec{
			UniformRepositorySpec: oci.UniformRepositorySpec{
				Type:   "directory",
				Scheme: "",
				Host:   "",
				Info:   "a/b",
			},
			ArtSpec: oci.ArtSpec{
				Repository: "c/d",
			},
		})

		CheckRef("oci::ghcr.io", &oci.RefSpec{
			UniformRepositorySpec: oci.UniformRepositorySpec{
				Type:   "oci",
				Scheme: "",
				Host:   "ghcr.io",
				Info:   "",
			},
			ArtSpec: oci.ArtSpec{
				Repository: "",
			},
		})
		CheckRef("/tmp/ctf//mandelsoft/test:v1", &oci.RefSpec{
			UniformRepositorySpec: oci.UniformRepositorySpec{
				Type:   "",
				Scheme: "",
				Host:   "",
				Info:   "/tmp/ctf",
			},
			ArtSpec: oci.ArtSpec{
				Repository: "mandelsoft/test",
				Tag:        &tag,
			},
		})
		CheckRef("/tmp/ctf", &oci.RefSpec{
			UniformRepositorySpec: oci.UniformRepositorySpec{
				Type:   "",
				Scheme: "",
				Host:   "",
				Info:   "/tmp/ctf",
			},
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
		CheckRepo("ghcr.io", &oci.UniformRepositorySpec{
			Host: "ghcr.io",
		})
		CheckRepo("https://ghcr.io", &oci.UniformRepositorySpec{
			Scheme: "https",
			Host:   "ghcr.io",
		})
		CheckRepo("alias", &oci.UniformRepositorySpec{
			Info: "alias",
		})
		CheckRepo("tar::a/b.tar", &oci.UniformRepositorySpec{
			Type: "tar",
			Info: "a/b.tar",
		})
		CheckRepo("a/b.tar", &oci.UniformRepositorySpec{
			Info: "a/b.tar",
		})
	})

})
