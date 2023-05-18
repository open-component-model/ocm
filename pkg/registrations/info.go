// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package registrations

import (
	"github.com/open-component-model/ocm/pkg/listformat"
)

type HandlerInfos []HandlerInfo

var _ listformat.ListElements = HandlerInfos(nil)

func (h HandlerInfos) Size() int {
	return len(h)
}

func (h HandlerInfos) Key(i int) string {
	return h[i].Name
}

func (h HandlerInfos) Description(i int) string {
	return h[i].ShortDesc + "\n" + h[i].Description
}

type HandlerInfo struct {
	Name        string
	ShortDesc   string
	Description string
}

func NewLeafHandlerInfo(short, desc string) HandlerInfos {
	return HandlerInfos{
		{
			ShortDesc:   short,
			Description: desc,
		},
	}
}
