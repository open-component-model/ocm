// Copyright 2022 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package main

import (
	"fmt"
	"io/ioutil"

	"github.com/open-component-model/ocm/pkg/contexts/config"
	"github.com/open-component-model/ocm/pkg/contexts/config/cpi"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/runtime"
)

const CFGFILE = "config.yaml"

type Target struct {
	updater cpi.Updater
	value   string
}

func NewTarget(ctx cpi.Context) *Target {
	return &Target{
		updater: cpi.NewUpdate(ctx),
	}
}

func (t *Target) SetValue(v string) {
	t.value = v
}

func (t *Target) GetValue() string {
	t.updater.Update(t)
	return t.value
}

////////////////////////////////////////////////////////////////////////////////

const TYPE = "mytype.config.mandelsoft.org"

func init() {
	cpi.RegisterConfigType(TYPE, cpi.NewConfigType(TYPE, &Config{}, "just provide a value for Target objects"))
}

type Config struct {
	runtime.ObjectVersionedType `json:",inline""`
	Value                       string `json:"value"`
}

func (c *Config) ApplyTo(context cpi.Context, i interface{}) error {
	if i == nil {
		return nil
	}
	t, ok := i.(*Target)
	if !ok {
		return cpi.ErrNoContext(TYPE)
	}
	t.SetValue(c.Value)
	return nil
}

var _ cpi.Config = (*Config)(nil)

func NewConfig(v string) *Config {
	return &Config{
		ObjectVersionedType: runtime.NewVersionedObjectType(TYPE),
		Value:               v,
	}
}

////////////////////////////////////////////////////////////////////////////////

func UsingConfigs() error {
	ctx := config.DefaultContext()

	target := NewTarget(ctx)

	err := ctx.ApplyConfig(NewConfig("hello world"), "explicit1")
	if err != nil {
		return errors.Wrapf(err, "cannot apply config 1")
	}

	fmt.Printf("value is %q\n", target.GetValue())

	err = ctx.ApplyConfig(NewConfig("hello universe"), "explicit2")
	if err != nil {
		return errors.Wrapf(err, "cannot apply config 2")
	}

	fmt.Printf("value is %q\n", target.GetValue())

	newtarget := NewTarget(ctx)
	fmt.Printf("value is %q\n", newtarget.GetValue())

	// now reading config from a central generic configuration file

	data, err := ioutil.ReadFile(CFGFILE)
	if err != nil {
		return errors.Wrapf(err, "cannot read configuration file %s", CFGFILE)
	}
	_, err = ctx.ApplyData(data, runtime.DefaultYAMLEncoding, CFGFILE)
	if err != nil {
		return errors.Wrapf(err, "cannot apply config data")
	}

	fmt.Printf("value is %q\n", newtarget.GetValue())

	return nil
}
