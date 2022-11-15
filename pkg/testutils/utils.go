// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package testutils

import (
	"fmt"
	"io"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func Close(c io.Closer, msg ...interface{}) {
	err := c.Close()
	if err != nil {
		switch len(msg) {
		case 0:
			ExpectWithOffset(1, err).To(Succeed())
		case 1:
			Fail(fmt.Sprintf("%s: %s", msg[0], err), 1)
		default:
			Fail(fmt.Sprintf("%s: %s", fmt.Sprintf(msg[0].(string), msg[1:]...), err), 1)
		}
	}
}

func Must[T any](o T, err error) T {
	ExpectWithOffset(1, err).To(Succeed())
	return o
}

func MustBeNonNil[T any](o T) T {
	ExpectWithOffset(1, o).NotTo(BeNil())
	return o
}

func MustBeSuccessful(err error) {
	ExpectWithOffset(1, err).To(Succeed())
}

func MustFailWithMessage(err error, msg string) {
	ExpectWithOffset(1, err).NotTo(BeNil())
	ExpectWithOffset(1, err.Error()).To(Equal(msg))
}
