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

package options

import (
	"github.com/gardener/ocm/cmds/ocm/clictx"
	"github.com/gardener/ocm/cmds/ocm/pkg/output/out"
	"github.com/spf13/pflag"
)

type OptionsProcessor func(Options) error

type Complete interface {
	Complete() error
}

type CompleteWithOutputContext interface {
	Complete(ctx out.Context) error
}

type CompleteWithCLIContext interface {
	Complete(ctx clictx.Context) error
}

type Usage interface {
	Usage() string
}

type Options interface {
	AddFlags(fs *pflag.FlagSet)
}

////////////////////////////////////////////////////////////////////////////////

func CompleteOptions(opt Options) error {
	if c, ok := opt.(Complete); ok {
		return c.Complete()
	}
	return nil
}

func CompleteOptionsWithOutputContext(ctx out.Context) OptionsProcessor {
	return func(opt Options) error {
		if c, ok := opt.(CompleteWithOutputContext); ok {
			return c.Complete(ctx)
		}
		if c, ok := opt.(Complete); ok {
			return c.Complete()
		}
		return nil
	}
}
func CompleteOptionsWithCLIContext(ctx clictx.Context) OptionsProcessor {
	return func(opt Options) error {
		if c, ok := opt.(CompleteWithCLIContext); ok {
			return c.Complete(ctx)
		}
		if c, ok := opt.(CompleteWithOutputContext); ok {
			return c.Complete(ctx)
		}
		if c, ok := opt.(Complete); ok {
			return c.Complete()
		}
		return nil
	}
}
