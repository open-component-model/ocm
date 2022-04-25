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

package scriptoption

import (
	"github.com/mandelsoft/vfs/pkg/vfs"
	cfgcpi "github.com/open-component-model/ocm/pkg/contexts/config/cpi"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/spf13/pflag"

	"github.com/open-component-model/ocm/cmds/ocm/clictx"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/options"
)

func From(o options.OptionSetProvider) *Option {
	var opt *Option
	o.AsOptionSet().Get(&opt)
	return opt
}

func New() *Option {
	return &Option{}
}

type Option struct {
	ScriptFile string
	Script     string
	ScriptData []byte
	FileSystem vfs.FileSystem
}

var _ options.OptionWithCLIContextCompleter = (*Option)(nil)

func (o *Option) AddFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&o.ScriptFile, "scriptFile", "s", "", "filename of transfer handler script")
	fs.StringVarP(&o.ScriptFile, "script", "", "", "config name of transfer handler script")
}

func (o *Option) Complete(ctx clictx.Context) error {
	o.FileSystem = ctx.FileSystem()
	if o.ScriptFile != "" && o.Script == "" {
		return errors.Newf("only one of --script or --scriptFile may be set")
	}
	if o.Script != "" {
		err := cfgcpi.NewUpdate(ctx.ConfigContext()).Update(o)
		if err != nil {
			return err
		}
		if o.ScriptData == nil {
			return errors.ErrUnknown("script", o.ScriptFile)
		}
	}
	if o.ScriptFile != "" {
		data, err := vfs.ReadFile(ctx.FileSystem(), o.ScriptFile)
		if err != nil {
			return errors.Wrapf(err, "invalid transfer script file")
		}
		o.ScriptData = data
	}
	if o.ScriptData == nil {
		o.Script = "default"
		err := cfgcpi.NewUpdate(ctx.ConfigContext()).Update(o)
		if err != nil {
			return err
		}
	}
	return nil
}

func (o *Option) Usage() string {
	s := `
It is possible to use a dedicated transfer script based on spiff.
The option <code>--scriptFile</code> can be used to specify this script
by a file name. With <code>--script</code> it can be taken from the 
CLI config using an entry of the following format:

<pre>
type: scripts.ocm.config.ocm.gardener.cloud
scripts:
  &lt;name>: 
    path: &lt;filepath> 
    script:
      &lt;scriptdata>
</pre>

Only one of the fields <code>path</code> or <code>script</code> can be used.

If no script option is given and the cli config defines a script <code>default</code>
this one is used.
`
	return s
}
