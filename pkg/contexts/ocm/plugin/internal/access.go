// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"github.com/open-component-model/ocm/pkg/contexts/credentials"
)

type AccessSpecInfo struct {
	Short      string
	MediaType  string
	Hint       string
	ConsumerId credentials.ConsumerIdentity
}
