// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package iotools

import (
	"io"

	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/errors"
)

////////////////////////////////////////////////////////////////////////////////

type additionalCloser struct {
	msg              []string
	reader           io.ReadCloser
	additionalCloser io.Closer
}

var _ io.ReadCloser = (*additionalCloser)(nil)

func AddCloser(reader io.ReadCloser, closer io.Closer, msg ...string) io.ReadCloser {
	return &additionalCloser{
		msg:              msg,
		reader:           reader,
		additionalCloser: closer,
	}
}

func (c *additionalCloser) Close() error {
	var list *errors.ErrorList
	if len(c.msg) == 0 {
		list = errors.ErrListf("close")
	} else {
		list = errors.ErrListf(c.msg[0], common.IterfaceSlice(c.msg[1:])...)
	}
	list.Add(c.reader.Close())
	list.Add(c.additionalCloser.Close())
	return list.Result()
}

func (c *additionalCloser) Read(p []byte) (n int, err error) {
	return c.reader.Read(p)
}
