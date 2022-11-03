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
