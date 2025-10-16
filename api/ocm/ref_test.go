package ocm_test

import (
	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/extensions/repositories/ocireg"
	"ocm.software/ocm/api/utils"
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

func Vers(t string) string {
	if t == "" {
		return t
	}
	return ":" + t
}

func CheckRef(ref, ut, scheme, h, us, c, uv, i string, th ...string) {
	var v *string
	if uv != "" {
		v = &uv
	}
	if len(th) == 0 && ut != "" {
		th = []string{ut}
	}
	spec, err := ocm.ParseRef(ref)
	Expect(err).WithOffset(1).To(Succeed())
	Expect(spec).WithOffset(1).To(Equal(ocm.RefSpec{
		UniformRepositorySpec: ocm.UniformRepositorySpec{
			Type:            ut,
			Scheme:          scheme,
			Host:            h,
			SubPath:         us,
			Info:            i,
			TypeHint:        utils.Optional(th...),
			CreateIfMissing: ref[0] == '+',
		},
		CompSpec: ocm.CompSpec{
			Component: c,
			Version:   v,
		},
	}))
}

var _ = Describe("ref parsing", func() {
	Context("file path refs", func() {
		t := "ctf"
		p := "file/path"
		c := "github.com/mandelsoft/ocm"
		v := "v1"

		Context("[+][<type>::][./][<file path>//<component id>[:<version>]", func() {
			for _, cm := range []string{"", "+"} {
				for _, ut := range []string{"", t} {
					for _, uf := range []string{"", "directory", "tar", "tgz"} {
						for _, up := range []string{p, "./" + p} {
							for _, uv := range []string{"", v, v + ".1.1", v + "-rc.1", v + "+65", v + ".1.2-rc.1", v + ".1.2+65"} {
								ref := cm + Type(FileFormat(ut, uf)) + up + "//" + c + Vers(uv)
								ut, uf, uv, up := ut, uf, uv, up

								// tests parsing of all permutations of
								// [+][<type>::][./]<file path>//<component id>[:<version>]
								It("parses ref "+ref, func() {
									if ut != "" || uf != "" {
										CheckRef(ref, FileType(ut, uf), "", "", "", c, uv, up, FileFormat(ut, uf))
									} else {
										CheckRef(ref, FileType(ut, uf), "", "", "", c, uv, up)
									}
								})
							}
						}
					}
				}
			}
		})
	})

	Context("json repo spec refs", func() {
		t := ocireg.Type
		s := "mandelsoft/cnudie"
		v := "v1"

		h := "ghcr.io"
		c := "github.com/mandelsoft/ocm"

		repospec := ocireg.NewRepositorySpec(h, &ocireg.ComponentRepositoryMeta{
			ComponentNameMapping: "",
			SubPath:              s,
		})
		jsonrepospec := string(Must(repospec.MarshalJSON()))

		Context("[<type>::][<json repo spec>//]<component id>[:<version>]", func() {
			for _, cm := range []string{"", "+"} {
				for _, ut := range []string{t, ""} {
					for _, uv := range []string{"", v, v + ".1.1", v + "-rc.1", v + "+65", v + ".1.2-rc.1", v + ".1.2+65"} {
						ref := cm + Type(ut) + jsonrepospec + "//" + c + Vers(uv)
						ut, uv := ut, uv

						// tests parsing of all permutations of
						// [<type>::][<json repo spec>//]<component id>[:<version>]
						It("parses ref "+ref, func() {
							CheckRef(ref, ut, "", "", "", c, uv, jsonrepospec)
						})
					}
				}
			}
		})

		It("fail if mismatch between type in ref (here, ctf) and type in json repo spec (here, OCIRegistry)", func() {
			ctx := ocm.New()

			ref := Must(ocm.ParseRef("ctf::{\"baseUrl\":\"ghcr.io\",\"subPath\":\"mandelsoft/cnudie\",\"type\":\"OCIRegistry\"}//github.com/mandelsoft/ocm:v1"))
			spec, err := ctx.MapUniformRepositorySpec(&ref.UniformRepositorySpec)
			Expect(spec).To(BeNil())
			Expect(err).ToNot(BeNil())
		})
	})

	Context("domain refs", func() {
		t := ocireg.Type
		s := "mandelsoft/cnudie"
		v := "v1"

		h := "ghcr.io"
		c := "github.com/mandelsoft/ocm"

		Context("[+][<type>::][<scheme>://]<domain>[:<port>][/<repository prefix>]//<component id>[:<version] - without info", func() {
			for _, cm := range []string{"", "+"} {
				for _, ut := range []string{t, ""} {
					for _, ush := range []string{"", "http", "https"} {
						for _, uh := range []string{h, h + ":3030"} {
							for _, us := range []string{"", s} {
								for _, uv := range []string{"", v, v + ".1.1", v + "-rc.1", v + "+65", v + ".1.2-rc.1", v + ".1.2+65"} {
									ref := cm + Type(ut) + Scheme(ush) + uh + Sub(us) + "//" + c + Vers(uv)
									ut, ush, uh, us, uv := ut, ush, uh, us, uv

									// tests parsing of all permutations of
									// [+][<type>::][scheme://]<domain>[:<port>][/<repository prefix>]//<component id>[:<version]
									It("parses ref "+ref, func() {
										CheckRef(ref, ut, ush, uh, us, c, uv, "")
									})
								}
							}
						}
					}
				}
			}
		})

		Context("host port refs", func() {
			t := ocireg.Type
			s := "mandelsoft/cnudie"
			v := "v1"

			h := "localhost"
			ip := "127.0.0.1"
			c := "github.com/mandelsoft/ocm"

			Context("[+][<type>::]<scheme>://<host>[:<port>][/<repository prefix>]//<component id>[:<version] - without info", func() {
				for _, cm := range []string{"", "+"} {
					for _, ut := range []string{t, ""} {
						for _, ush := range []string{"http", "https"} {
							for _, uh := range []string{h, h + ":3030", ip, ip + ":3030"} {
								for _, us := range []string{"", s} {
									for _, uv := range []string{"", v, v + ".1.1", v + "-rc.1", v + "+65", v + ".1.2-rc.1", v + ".1.2+65"} {
										ref := cm + Type(ut) + Scheme(ush) + uh + Sub(us) + "//" + c + Vers(uv)
										ut, ush, uh, us, uv := ut, ush, uh, us, uv

										// tests parsing of all permutations of
										// [+][<type>::]scheme://<host>[:<port>][/<repository prefix>]//<component id>[:<version]
										It("parses ref "+ref, func() {
											CheckRef(ref, ut, ush, uh, us, c, uv, "")
										})
									}
								}
							}
						}
					}
				}
			})

			Context("[+][<type>::][<scheme>://]<host>:<port>[/<repository prefix>]//<component id>[:<version] - without info", func() {
				for _, cm := range []string{"", "+"} {
					for _, ut := range []string{t, ""} {
						for _, ush := range []string{"", "http", "https"} {
							for _, uh := range []string{h + ":3030", ip + ":3030"} {
								for _, us := range []string{"", s} {
									for _, uv := range []string{"", v, v + ".1.1", v + "-rc.1", v + "+65", v + ".1.2-rc.1", v + ".1.2+65"} {
										ref := cm + Type(ut) + Scheme(ush) + uh + Sub(us) + "//" + c + Vers(uv)
										ut, ush, uh, us, uv := ut, ush, uh, us, uv

										// tests parsing of all permutations of
										// [+][<type>::][scheme://]<host>:<port>[/<repository prefix>]//<component id>[:<version]
										It("parses ref "+ref, func() {
											CheckRef(ref, ut, ush, uh, us, c, uv, "")
										})
									}
								}
							}
						}
					}
				}
			})
		})

		It("scheme", func() {
			CheckRef("OCIRegistry::http://ghcr.io:80/repository//acme.org/component:1.0.0", "OCIRegistry", "http", "ghcr.io:80", "repository", "acme.org/component", "1.0.0", "")
		})

		It("info", func() {
			for _, ut := range []string{"", t} {
				CheckRef(Type(ut)+"{}", ut, "", "", "", "", "", "{}")
			}
		})

		It("info+comp", func() {
			for _, ut := range []string{"", t} {
				for _, uv := range []string{"", v} {
					CheckRef(Type(ut)+"{}"+"//"+c+Vers(uv), ut, "", "", "", c, uv, "{}")
				}
			}
		})

		It("dir ref", func() {
			CheckRef("+ctf+directory::./file//bla.blob/comp", "ctf", "", "", "", "bla.blob/comp", "", "./file", "ctf+directory")
			CheckRef("ctf+directory::./file//bla.blob/comp", "ctf", "", "", "", "bla.blob/comp", "", "./file", "ctf+directory")
			CheckRef("directory::./file//bla.blob/comp", "directory", "", "", "", "bla.blob/comp", "", "./file")
			CheckRef("directory::file//bla.blob/comp", "directory", "", "", "", "bla.blob/comp", "", "file")
			CheckRef("directory::./file.io//bla.blob/comp", "directory", "", "", "", "bla.blob/comp", "", "./file.io")
			CheckRef("any::file.io//bla.blob/comp", "any", "", "file.io", "", "bla.blob/comp", "", "")
		})

		It("dedicated test case", func() {
			ref := Must(ocm.ParseRef("OCIRegistry::73555000100900003892.dev.dockersrv.repositories.sapcloud.cn/dev/v1//github.wdf.sap.corp/kubernetes/landscape-setup-dependencies:0.3797.0"))

			v := "0.3797.0"
			Expect(ref).To(Equal(ocm.RefSpec{
				UniformRepositorySpec: ocm.UniformRepositorySpec{
					Type:            "OCIRegistry",
					Host:            "73555000100900003892.dev.dockersrv.repositories.sapcloud.cn",
					SubPath:         "dev/v1",
					Info:            "",
					CreateIfMissing: false,
					TypeHint:        "OCIRegistry",
				},
				CompSpec: ocm.CompSpec{
					Component: "github.wdf.sap.corp/kubernetes/landscape-setup-dependencies",
					Version:   &v,
				},
			}))
		})
	})

	Context("json repository spec ref", func() {
		It("type in ref", func() {
			ctx := ocm.New()

			ref := Must(ocm.ParseRef("+oci::{\"baseUrl\": \"example.com\"}//github.com/mandelsoft/ocm:v1"))
			spec := Must(ctx.MapUniformRepositorySpec(&ref.UniformRepositorySpec))
			repo := Must(spec.Repository(ctx, nil))
			_ = repo
		})
		It("type in json repo spec", func() {
			ctx := ocm.New()

			ref := Must(ocm.ParseRef("{\"type\":\"OCIRegistry\", \"baseUrl\": \"example.com\"}//github.com/mandelsoft/ocm:v1"))
			spec := Must(ctx.MapUniformRepositorySpec(&ref.UniformRepositorySpec))
			repo := Must(spec.Repository(ctx, nil))
			_ = repo
		})
		It("type in ref and json repo spec", func() {
			ctx := ocm.New()

			ref := Must(ocm.ParseRef("OCIRegistry::{\"type\":\"OCIRegistry\", \"baseUrl\": \"example.com//test\"}//github.com/mandelsoft/ocm:v1"))
			spec := Must(ctx.MapUniformRepositorySpec(&ref.UniformRepositorySpec))
			repo := Must(spec.Repository(ctx, nil))
			_ = repo
		})
	})

	Context("map to spec", func() {
		It("handles localhost", func() {
			ctx := ocm.New()

			ref := Must(ocm.ParseRef("OCIRegistry::http://localhost:80/test//github.vom/mandelsoft/test"))
			spec := Must(ctx.MapUniformRepositorySpec(&ref.UniformRepositorySpec))
			Expect(spec).To(Equal(ocireg.NewRepositorySpec("http://localhost:80", ocireg.NewComponentRepositoryMeta("test", ""))))
			repo := Must(spec.Repository(ctx, nil))
			specFromRepo := repo.GetSpecification()
			Expect(spec).To(Equal(specFromRepo))
		})
		It("handles localhost without scheme", func() {
			ctx := ocm.New()

			ref := Must(ocm.ParseRef("OCIRegistry::localhost:80/test//github.vom/mandelsoft/test"))
			spec := Must(ctx.MapUniformRepositorySpec(&ref.UniformRepositorySpec))
			Expect(spec).To(Equal(ocireg.NewRepositorySpec("localhost:80", ocireg.NewComponentRepositoryMeta("test", ""))))
			repo := Must(spec.Repository(ctx, nil))
			specFromRepo := repo.GetSpecification()
			Expect(spec).To(Equal(specFromRepo))
		})
	})
})
