// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package simplelistmerge

import (
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/labelmergehandler/handlers/simplemapmerge"
)

type Mode = simplemapmerge.Mode

const (
	MODE_NONE    = simplemapmerge.MODE_NONE
	MODE_LOCAL   = simplemapmerge.MODE_LOCAL
	MODE_INBOUND = simplemapmerge.MODE_INBOUND
)

func NewConfig(field string, overwrite Mode) *Config {
	return &Config{
		KeyField: field,
		Config:   *simplemapmerge.NewConfig(overwrite),
	}
}

type Config struct {
	KeyField string `json:"keyField"`
	simplemapmerge.Config
}

func (c *Config) Complete(ctx cpi.Context) error {
	err := c.Config.Complete(ctx)
	if err != nil {
		return err
	}
	if c.KeyField == "" {
		c.KeyField = "name"
	}
	return nil
}
