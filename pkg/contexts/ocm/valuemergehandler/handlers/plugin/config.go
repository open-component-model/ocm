// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package plugin

import (
	"encoding/json"

	"github.com/open-component-model/ocm/pkg/contexts/ocm/valuemergehandler"
)

type Config struct {
	json.RawMessage
}

var _ valuemergehandler.Config = (*Config)(nil)

func (c Config) Complete(valuemergehandler.Context) error {
	return nil
}
