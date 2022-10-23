//  SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
//  SPDX-License-Identifier: Apache-2.0

package flagsets

type Config = map[string]interface{}

type ConfigAdder func(options ConfigOptions, config Config) error

type ConfigHandler interface {
	ApplyConfig(options ConfigOptions, config Config) error
}

type ConfigOptionTypeSetHandler interface {
	ConfigOptionTypeSet
	ConfigHandler
}
type configOptionTypeSetHandler struct {
	adder ConfigAdder
	ConfigOptionTypeSet
}

func NewConfigOptionTypeSetHandler(name string, adder ConfigAdder, types ...ConfigOptionType) ConfigOptionTypeSetHandler {
	return &configOptionTypeSetHandler{
		adder:               adder,
		ConfigOptionTypeSet: NewConfigOptionSet(name, types...),
	}
}

func (c *configOptionTypeSetHandler) ApplyConfig(options ConfigOptions, config Config) error {
	if c.adder == nil {
		return nil
	}
	return c.adder(options, config)
}

type nopConfigHandler struct{}

var NopConfigHandler = NewNopConfigHandler()

func NewNopConfigHandler() ConfigHandler {
	return &nopConfigHandler{}
}

func (c *nopConfigHandler) ApplyConfig(options ConfigOptions, config Config) error {
	return nil
}
