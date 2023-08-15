// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package simplemapmerge

import (
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/errors"
)

type Mode string

func (m Mode) String() string {
	return string(m)
}

const (
	MODE_NONE    = Mode("none")
	MODE_LOCAL   = Mode("local")
	MODE_INBOUND = Mode("inbound")
)

func NewConfig(overwrite Mode) *Config {
	return &Config{
		Overwrite: overwrite,
	}
}

type Config struct {
	Overwrite Mode `json:"overwrite"`
}

func (c *Config) Complete(ctx cpi.Context) error {
	switch c.Overwrite {
	case MODE_NONE, MODE_LOCAL, MODE_INBOUND:
	case "":
		c.Overwrite = MODE_NONE
	default:
		return errors.ErrInvalid("merge overwrite mode", string(c.Overwrite))
	}
	return nil
}
