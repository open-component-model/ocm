// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package testhelper

import (
	. "github.com/onsi/gomega"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/oci/repositories/ctf"
	"github.com/open-component-model/ocm/pkg/env/builder"
)

func FakeOCIRepo(env *builder.Builder, path string, host string) {
	spec, err := ctf.NewRepositorySpec(accessobj.ACC_READONLY, path, accessio.PathFileSystem(env.FileSystem()))
	ExpectWithOffset(1, err).To(Succeed())
	env.OCIContext().SetAlias(host, spec)
}
