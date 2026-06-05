package accessmethods_test

import (
	"fmt"
	"reflect"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/cpi"

	_ "ocm.software/ocm/api/ocm/extensions/accessmethods"

	"ocm.software/ocm/api/ocm/extensions/accessmethods/git"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/github"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/helm"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/localblob"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/maven"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/npm"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/ociartifact"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/ociblob"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/wget"
)

type aliasCase struct {
	method   string
	alias    string
	body     string
	wantType func(cpi.AccessSpec) bool
	is       func(cpi.AccessSpec) bool
}

func concreteIs(proto cpi.AccessSpec) func(cpi.AccessSpec) bool {
	want := reflect.TypeOf(proto)
	return func(s cpi.AccessSpec) bool {
		return s != nil && reflect.TypeOf(s) == want
	}
}

func cases() []aliasCase {
	var c []aliasCase

	add := func(method, body string, wantType func(cpi.AccessSpec) bool, is func(cpi.AccessSpec) bool, aliases ...string) {
		for _, a := range aliases {
			c = append(c, aliasCase{method, a, body, wantType, is})
		}
	}

	add("localblob",
		`"localReference":"abc","mediaType":"text/plain"`,
		concreteIs((*localblob.AccessSpec)(nil)), localblob.Is,
		localblob.Type, localblob.TypeV1, localblob.UpperType, localblob.UpperTypeV1,
	)
	add("ociartifact",
		`"imageReference":"ghcr.io/x/y:1"`,
		concreteIs((*ociartifact.AccessSpec)(nil)), ociartifact.Is,
		ociartifact.Type, ociartifact.TypeV1,
		ociartifact.LegacyType, ociartifact.LegacyTypeV1,
		ociartifact.LegacyType2, ociartifact.LegacyType2V1,
	)
	add("ociblob",
		`"ref":"ghcr.io/x/y","digest":"sha256:0000000000000000000000000000000000000000000000000000000000000000","mediaType":"application/octet-stream","size":1`,
		concreteIs((*ociblob.AccessSpec)(nil)), nil,
		ociblob.Type, ociblob.TypeV1, ociblob.UpperType, ociblob.UpperTypeV1,
	)
	add("helm",
		`"helmChart":"chart:1.0.0","helmRepository":"https://charts.example.com"`,
		concreteIs((*helm.AccessSpec)(nil)), nil,
		helm.Type, helm.TypeV1, helm.UpperType, helm.UpperTypeV1,
	)
	add("github",
		`"repoUrl":"https://github.com/x/y","commit":"0000000000000000000000000000000000000000"`,
		concreteIs((*github.AccessSpec)(nil)), github.Is,
		github.Type, github.TypeV1,
		github.LegacyType, github.LegacyTypeV1,
		github.UpperType, github.UpperTypeV1,
	)
	add("git",
		`"repository":"https://example.com/x.git"`,
		concreteIs((*git.AccessSpec)(nil)), nil,
		git.Type, git.TypeV1Alpha1, git.UpperType, git.UpperTypeV1Alpha1,
	)
	add("npm",
		`"registry":"https://registry.npmjs.org","package":"x","version":"1.0.0"`,
		concreteIs((*npm.AccessSpec)(nil)), nil,
		npm.Type, npm.TypeV1, npm.UpperType, npm.UpperTypeV1,
	)
	add("wget",
		`"URL":"https://example.com/x.tgz"`,
		concreteIs((*wget.AccessSpec)(nil)), wget.Is,
		wget.Type, wget.TypeV1, wget.UpperType, wget.UpperTypeV1,
	)
	add("maven",
		`"repoUrl":"https://repo.example.com","groupId":"com.example","artifactId":"x","version":"1.0.0"`,
		concreteIs((*maven.AccessSpec)(nil)), nil,
		maven.Type, maven.TypeV1, maven.UpperType, maven.UpperTypeV1,
	)

	return c
}

var _ = Describe("Access method alias surface", func() {
	var ctx ocm.Context

	BeforeEach(func() { ctx = ocm.New() })
	AfterEach(func() { ctx.Finalize() })

	for _, tc := range cases() {
		tc := tc
		label := fmt.Sprintf("%s/%s", tc.method, tc.alias)
		It(label, func() {
			data := []byte(fmt.Sprintf(`{"type":%q,%s}`, tc.alias, tc.body))

			spec, err := ctx.AccessSpecForConfig(data, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(spec).NotTo(BeNil())
			Expect(tc.wantType(spec)).To(BeTrue(), "wrong concrete type: %T", spec)

			if tc.is != nil {
				Expect(tc.is(spec)).To(BeTrue(), "Is() returned false — dispatch hole, see #1979")
			}
		})
	}
})
