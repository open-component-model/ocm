// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package accessmethods_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type AccessSpec struct {
	Type      string `json:"type"`
	Path      string `json:"path"`
	MediaType string `json:"mediaType"`
}

func TestConfig(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Access Methods Plugin Test Suite")
}
