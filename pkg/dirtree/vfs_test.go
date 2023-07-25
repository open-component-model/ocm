// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package dirtree_test

import (
	"archive/tar"
	"encoding/hex"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/v2/pkg/testutils"

	"github.com/mandelsoft/logging"

	"github.com/open-component-model/ocm/v2/pkg/dirtree"
	"github.com/open-component-model/ocm/v2/pkg/env"
)

var _ = Describe("file system", func() {
	var tenv *env.Environment
	lctx := logging.NewDefault()
	lctx.SetDefaultLevel(logging.TraceLevel)
	ctx := dirtree.DefaultContext(lctx)

	BeforeEach(func() {
		tenv = env.NewEnvironment(env.ModifiableTestData())
	})

	AfterEach(func() {
		tenv.Cleanup()
	})

	It("creates treehash", func() {
		ctx.Logger().Trace("dirhash")
		n := Must(dirtree.NewVFSDirNode(ctx, "/testdata/fs", tenv))
		Expect(hex.EncodeToString(n.Hash())).To(Equal("7392f1050d5e9efc378f4b052f307d3613285cf6"))
	})

	It("creates tarhash", func() {
		ctx.Logger().Trace("tarhash")
		file := Must(tenv.Open("/testdata/fs.tar"))
		defer file.Close()
		tr := tar.NewReader(file)
		n := Must(dirtree.NewTarDirNode(ctx, tr))
		Expect(hex.EncodeToString(n.Hash())).To(Equal("7392f1050d5e9efc378f4b052f307d3613285cf6"))
	})
})
