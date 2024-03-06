// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package oci_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/oci/repositories/ctf"
	"github.com/open-component-model/ocm/pkg/runtime"
	. "github.com/open-component-model/ocm/pkg/testutils"

	godigest "github.com/opencontainers/go-digest"

	"github.com/open-component-model/ocm/pkg/contexts/oci"
	"github.com/open-component-model/ocm/pkg/contexts/oci/repositories/ocireg"
)

func Type(t string) string {
	if t == "" {
		return t
	}
	return t + "::"
}
func FileFormat(t, f string) string {
	if t == "" {
		return f
	}
	if f == "" {
		return t
	}
	return t + "+" + f
}
func FileType(t, f string) string {
	if t != "" {
		return t
	} else {
		return f
	}
}
func Scheme(s string) string {
	if s == "" {
		return s
	}
	return s + "://"
}
func Sub(t string) string {
	if t == "" {
		return t
	}
	return "/" + t
}
func Vers(t, d string) string {
	if t == "" && d == "" {
		return ""
	}
	if t == "" {
		return "@" + d
	}
	if d == "" {
		return ":" + t
	}
	return ":" + t + "@" + d
}

func Dig(b []byte) *godigest.Digest {
	if len(b) == 0 {
		return nil
	}
	s := godigest.Digest(b)
	return &s
}

func Pointer(b []byte) *string {
	if len(b) == 0 {
		return nil
	}
	s := string(b)
	return &s
}

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
	digest := godigest.Digest("sha256:3d05e105e350edf5be64fe356f4906dd3f9bf442a279e4142db9879bba8e677a")
	tag := "v1"

	ghcr := oci.UniformRepositorySpec{Host: "ghcr.io"}
	docker := oci.UniformRepositorySpec{Host: "docker.io"}

	Context("file path refs", func() {
		t := "ctf"
		p := "file/path"
		r := "github.com/mandelsoft/ocm"
		v := "v1"
		d := "sha256:3d05e105e350edf5be64fe356f4906dd3f9bf442a279e4142db9879bba8e677a"

		Context("[+][<type>::][./][<file path>//<component id>[:<version>]", func() {
			for _, cm := range []string{"", "+"} {
				for _, ut := range []string{"", t} {
					for _, uf := range []string{"", "directory", "tar", "tgz"} {
						for _, up := range []string{p, "./" + p} {
							for _, uv := range []string{"", v, v + ".1.1", v + "-rc.1", v + ".1.2-rc.1"} {
								for _, ud := range []string{"", d} {
									ref := cm + Type(FileFormat(ut, uf)) + up + "//" + r + Vers(uv, ud)
									ut, uf, uv, up, ud := ut, uf, uv, up, ud

									// tests parsing of all permutations of
									// [+][<type>::][./][<file path>//<repository>[:<tag>][@<digest>]
									It("parses ref "+ref, func() {
										CheckRef(ref, &oci.RefSpec{
											UniformRepositorySpec: oci.UniformRepositorySpec{
												Type:            FileType(ut, uf),
												Scheme:          "",
												Host:            "",
												Info:            up,
												CreateIfMissing: ref[0] == '+',
												TypeHint:        FileFormat(ut, uf),
											},
											ArtSpec: oci.ArtSpec{
												Repository: r,
												Tag:        Pointer([]byte(uv)),
												Digest:     Dig([]byte(ud)),
											},
										})
									})
								}
							}
						}
					}
				}
			}
		})
	})

	Context("domain refs", func() {
		t := "oci"
		h := "ghcr.io"
		r := "github.com/mandelsoft/ocm"
		v := "v1"
		d := "sha256:3d05e105e350edf5be64fe356f4906dd3f9bf442a279e4142db9879bba8e677a"

		// Notice that the file formats (directory, tar, tgz) CAN BE PARSED in this notation, BUT for non file based
		// implementations like oci, this information is not used.
		Context("[<type>::][<scheme>:://]<domain>[:<port>/]<repository>[:<tag>][@<digest>]", func() {
			for _, cm := range []string{"", "+"} {
				for _, ut := range []string{"", t} {
					for _, uf := range []string{"", "directory", "tar", "tgz"} {
						for _, ush := range []string{"", "http", "https"} {
							for _, uh := range []string{h, h + ":3030"} {
								for _, uv := range []string{"", v, v + ".1.1", v + "-rc.1", v + ".1.2-rc.1"} {
									for _, ud := range []string{"", d} {
										ref := cm + Type(FileFormat(ut, uf)) + Scheme(ush) + uh + "/" + r + Vers(uv, ud)
										ut, uf, ush, uh, uv, ud := ut, uf, ush, uh, uv, ud

										// tests parsing of all permutations of
										// [<type>::][<scheme>:://]<domain>[:<port>/]<repository>[:<tag>][@<digest>]
										It("parses ref "+ref, func() {
											CheckRef(ref, &oci.RefSpec{
												UniformRepositorySpec: oci.UniformRepositorySpec{
													Type:            FileType(ut, uf),
													Scheme:          ush,
													Host:            uh,
													Info:            "",
													CreateIfMissing: ref[0] == '+',
													TypeHint:        FileFormat(ut, uf),
												},
												ArtSpec: oci.ArtSpec{
													Repository: r,
													Tag:        Pointer([]byte(uv)),
													Digest:     Dig([]byte(ud)),
												},
											})
										})
									}
								}
							}
						}
					}
				}
			}
		})
	})

	Context("host port refs", func() {
		t := "oci"
		h := "localhost"
		r := "github.com/mandelsoft/ocm"
		v := "v1"
		d := "sha256:3d05e105e350edf5be64fe356f4906dd3f9bf442a279e4142db9879bba8e677a"

		// localhost (with and without port) (and other host names) are a special case since these are not formally
		// valid domains
		// the combination of this test and the test below test parsing of all permutations of
		// [<type>::][<scheme>://]<host>:<port>/<repository>[:<tag>][@<digest>]
		Context("[<type>::]<scheme>://<host>:<port>/<repository>[:<tag>][@<digest>]", func() {
			for _, cm := range []string{"", "+"} {
				for _, ut := range []string{"", t} {
					for _, uf := range []string{"", "directory", "tar", "tgz"} {
						for _, ush := range []string{"", "http", "https"} {
							for _, uh := range []string{h + ":3030"} {
								for _, uv := range []string{"", v, v + ".1.1", v + "-rc.1", v + ".1.2-rc.1"} {
									for _, ud := range []string{"", d} {
										ref := cm + Type(FileFormat(ut, uf)) + Scheme(ush) + uh + "/" + r + Vers(uv, ud)
										ut, uf, ush, uh, uv, ud := ut, uf, ush, uh, uv, ud

										// tests parsing of all permutations of
										// [<type>::]<scheme>://<domain>:<port>/<repository>[:<tag>][@<digest>]
										It("parses ref "+ref, func() {
											CheckRef(ref, &oci.RefSpec{
												UniformRepositorySpec: oci.UniformRepositorySpec{
													Type:            FileType(ut, uf),
													Scheme:          ush,
													Host:            uh,
													Info:            "",
													CreateIfMissing: ref[0] == '+',
													TypeHint:        FileFormat(ut, uf),
												},
												ArtSpec: oci.ArtSpec{
													Repository: r,
													Tag:        Pointer([]byte(uv)),
													Digest:     Dig([]byte(ud)),
												},
											})
										})
									}
								}
							}
						}
					}
				}
			}
		})
		//Context("[<type>::]<host>:<port>/<repository>[:<tag>][@<digest>]", func() {
		//	for _, cm := range []string{"", "+"} {
		//		for _, ut := range []string{"", t} {
		//			for _, uf := range []string{"", "directory", "tar", "tgz"} {
		//				for _, uh := range []string{h + ":3030"} {
		//					for _, uv := range []string{"", v, v + ".1.1", v + "-rc.1", v + ".1.2-rc.1"} {
		//						for _, ud := range []string{"", d} {
		//							ref := cm + Type(FileFormat(ut, uf)) + uh + "/" + r + Vers(uv, ud)
		//							ut, uf, uh, uv, ud := ut, uf, uh, uv, ud
		//
		//							// tests parsing of all permutations of
		//							// [<type>::]<host>:<port>/<repository>[:<tag>][@<digest>]
		//							It("parses ref "+ref, func() {
		//								CheckRef(ref, &oci.RefSpec{
		//									UniformRepositorySpec: oci.UniformRepositorySpec{
		//										Type:            FileType(ut, uf),
		//										Scheme:          "",
		//										Host:            "",
		//										Info:            uh,
		//										CreateIfMissing: ref[0] == '+',
		//										TypeHint:        FileFormat(ut, uf),
		//									},
		//									ArtSpec: oci.ArtSpec{
		//										Repository: r,
		//										Tag:        Pointer([]byte(uv)),
		//										Digest:     Dig([]byte(ud)),
		//									},
		//								})
		//							})
		//						}
		//					}
		//				}
		//			}
		//		}
		//	}
		//})
	})

	Context("json repo spec refs", func() {
		t := "oci"
		h := "ghcr.io"
		r := "github.com/mandelsoft/ocm"
		v := "v1"
		d := "sha256:3d05e105e350edf5be64fe356f4906dd3f9bf442a279e4142db9879bba8e677a"

		repospec := ocireg.NewRepositorySpec(h)
		jsonrepospec := string(Must(runtime.DefaultJSONEncoding.Marshal(repospec)))

		// Notice that the file formats (directory, tar, tgz) CAN BE PARSED in this notation, BUT for non file based
		// implementations like oci, this information is not used.
		Context("[<type>::][<json repo spec>//]<repository>[:<tag>][@<digest>]", func() {
			for _, cm := range []string{"", "+"} {
				for _, ut := range []string{"", t} {
					for _, uf := range []string{"", "directory", "tar", "tgz"} {
						for _, uv := range []string{"", v, v + ".1.1", v + "-rc.1", v + ".1.2-rc.1"} {
							for _, ud := range []string{"", d} {
								ref := cm + Type(ut) + jsonrepospec + "//" + r + Vers(uv, ud)
								ut, uv, ud := ut, uv, ud

								// tests parsing of all permutations of
								// [<type>::][<json repo spec>//]<repository>[:<tag>][@<digest>]
								It("parses ref "+ref, func() {
									CheckRef(ref, &oci.RefSpec{
										UniformRepositorySpec: oci.UniformRepositorySpec{
											Type:            FileType(ut, uf),
											Scheme:          "",
											Host:            "",
											Info:            jsonrepospec,
											CreateIfMissing: ref[0] == '+',
											TypeHint:        FileFormat(ut, uf),
										},
										ArtSpec: oci.ArtSpec{
											Repository: r,
											Tag:        Pointer([]byte(uv)),
											Digest:     Dig([]byte(ud)),
										},
									})
								})
							}
						}
					}
				}
			}
		})
	})

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
				Type:     "type",
				Scheme:   "https",
				Host:     "ghcr.io",
				Info:     "",
				TypeHint: "type",
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
				Type:     "directory",
				Scheme:   "",
				Host:     "",
				Info:     "a/b",
				TypeHint: "directory",
			},
			ArtSpec: oci.ArtSpec{
				Repository: "",
			},
		})
		CheckRef("ctf+directory::a/b", &oci.RefSpec{
			UniformRepositorySpec: oci.UniformRepositorySpec{
				Type:     "ctf",
				Scheme:   "",
				Host:     "",
				Info:     "a/b",
				TypeHint: "ctf+directory",
			},
			ArtSpec: oci.ArtSpec{
				Repository: "",
			},
		})
		CheckRef("+ctf+directory::a/b", &oci.RefSpec{
			UniformRepositorySpec: oci.UniformRepositorySpec{
				Type:            "ctf",
				Scheme:          "",
				Host:            "",
				Info:            "a/b",
				CreateIfMissing: true,
				TypeHint:        "ctf+directory",
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
				Type:     "directory",
				Scheme:   "",
				Host:     "",
				Info:     "a/b",
				TypeHint: "directory",
			},
			ArtSpec: oci.ArtSpec{
				Repository: "c/d",
			},
		})

		CheckRef("oci::ghcr.io", &oci.RefSpec{
			UniformRepositorySpec: oci.UniformRepositorySpec{
				Type:     "oci",
				Scheme:   "",
				Host:     "ghcr.io",
				Info:     "",
				TypeHint: "oci",
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

	It("json spec", func() {
		ctx := oci.New()

		tag := "1.0.0"
		CheckRef("OCIRegistry::{\"baseUrl\": \"test.com\"}//repo:1.0.0", &oci.RefSpec{
			UniformRepositorySpec: oci.UniformRepositorySpec{
				Type:     "OCIRegistry",
				Scheme:   "",
				Host:     "",
				Info:     "{\"baseUrl\": \"test.com\"}",
				TypeHint: "OCIRegistry",
			},
			ArtSpec: oci.ArtSpec{
				Repository: "repo",
				Tag:        &tag,
			},
		})
		ref := Must(oci.ParseRef("OCIRegistry::{\"baseUrl\": \"test.com\"}//repo:1.0.0"))
		spec := Must(ctx.MapUniformRepositorySpec(&ref.UniformRepositorySpec))
		repo := Must(spec.Repository(ctx, nil))
		_ = repo
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
			Type:     "tar",
			Info:     "a/b.tar",
			TypeHint: "tar",
		})
		CheckRepo("a/b.tar", &oci.UniformRepositorySpec{
			Info: "a/b.tar",
		})
	})
	It("localhost", func() {
		ctx := oci.New()
		// port is necessary here, otherwise it is ambiguous with dockerhub reference (localhost/test:1.0.0 could be
		// an artifact stored on duckerhub)
		ref := Must(oci.ParseRef("localhost:80/test:1.0.0"))
		spec := Must(ctx.MapUniformRepositorySpec(&ref.UniformRepositorySpec))
		Expect(spec).To(Equal(ocireg.NewRepositorySpec("localhost:80")))
	})
	It("localhost with unambiguous separator and without port", func() {
		ctx := oci.New()
		ref := Must(oci.ParseRef("localhost//test:1.0.0"))
		spec := Must(ctx.MapUniformRepositorySpec(&ref.UniformRepositorySpec))
		Expect(spec).To(Equal(ocireg.NewRepositorySpec("localhost")))
	})
	It("localhost with unambiguous separator", func() {
		ctx := oci.New()
		ref := Must(oci.ParseRef("localhost:80//test:1.0.0"))
		spec := Must(ctx.MapUniformRepositorySpec(&ref.UniformRepositorySpec))
		Expect(spec).To(Equal(ocireg.NewRepositorySpec("localhost:80")))
	})
	It("scheme://localhost:port//repository:version", func() {
		ctx := oci.New()
		ref := Must(oci.ParseRef("http://localhost:80//test:1.0.0"))
		spec := Must(ctx.MapUniformRepositorySpec(&ref.UniformRepositorySpec))
		Expect(spec).To(Equal(ocireg.NewRepositorySpec("http://localhost:80")))
	})
	It("scheme://localhost:port/repository:version", func() {
		ctx := oci.New()
		ref := Must(oci.ParseRef("http://localhost:80/test:1.0.0"))
		spec := Must(ctx.MapUniformRepositorySpec(&ref.UniformRepositorySpec))
		Expect(spec).To(Equal(ocireg.NewRepositorySpec("http://localhost:80")))
	})
	It("ctf with create", func() {
		ctx := oci.New()
		ref := Must(oci.ParseRef("+ctf+directory::./file/path//github.com/mandelsoft/ocm"))
		spec := Must(ctx.MapUniformRepositorySpec(&ref.UniformRepositorySpec))
		Expect(spec).To(Equal(Must(ctf.NewRepositorySpec(accessobj.ACC_CREATE, "./file/path", accessio.FormatDirectory))))
	})
	It("ctf without create", func() {
		ctx := oci.New()

		ref := Must(oci.ParseRepo("ctf+directory::./file/path"))
		spec := Must(ctx.MapUniformRepositorySpec(&ref))
		Expect(spec).To(Equal(Must(ctf.NewRepositorySpec(accessobj.ACC_WRITABLE, "./file/path"))))
	})
})
