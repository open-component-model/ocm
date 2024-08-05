package main

import (
	"fmt"
	"io/ioutil"

	"github.com/mandelsoft/goutils/errors"

	"ocm.software/ocm/api/config"
	"ocm.software/ocm/api/config/cpi"
	"ocm.software/ocm/api/utils/runtime"
)

const CFGFILE = "examples/lib/config3/config.yaml"

type Target struct {
	updater cpi.Updater
	value   string
}

func NewTarget(ctx cpi.Context) *Target {
	t := &Target{}
	t.updater = cpi.NewUpdater(ctx, t)
	return t
}

func (t *Target) SetValue(v string) {
	t.value = v
}

func (t *Target) GetValue() string {
	t.updater.Update()
	return t.value
}

////////////////////////////////////////////////////////////////////////////////

const TYPE = "mytype.config.mandelsoft.org"

func init() {
	cpi.RegisterConfigType(cpi.NewConfigType[*Config](TYPE, "just provide a value for Target objects"))
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
		ObjectVersionedType: runtime.NewVersionedTypedObject(TYPE),
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
