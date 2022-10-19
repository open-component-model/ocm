// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package testhelper

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"

	"github.com/open-component-model/ocm/pkg/contexts/oci/repositories/artefactset"
)

func TestForAllFormats(msg string, f func(fmt string)) {
	DescribeTable(fmt.Sprintf("%s: structure format handling", msg), f,
		Entry("OCM format", artefactset.FORMAT_OCM),
		Entry("OCI format", artefactset.FORMAT_OCI),
	)
}

func FTestForAllFormats(msg string, f func(fmt string)) {
	FDescribeTable(fmt.Sprintf("%s: structure format handling", msg), f,
		Entry("OCM format", artefactset.FORMAT_OCM),
		Entry("OCI format", artefactset.FORMAT_OCI),
	)
}
