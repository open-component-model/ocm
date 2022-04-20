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

package common

import (
	"github.com/open-component-model/ocm/cmds/ocm/clictx"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/options"
	"github.com/open-component-model/ocm/pkg/contexts/oci"
)

type OptionCompleter interface {
	CompleteWithSession(ctx clictx.OCI, session oci.Session) error
}

func CompleteOptionsWithContext(ctx clictx.Context, session oci.Session) options.OptionsProcessor {
	return func(opt options.Options) error {
		if c, ok := opt.(OptionCompleter); ok {
			return c.CompleteWithSession(ctx.OCI(), session)
		}
		if c, ok := opt.(options.CompleteWithCLIContext); ok {
			return c.Complete(ctx)
		}
		if c, ok := opt.(options.Complete); ok {
			return c.Complete()
		}
		return nil
	}
}
