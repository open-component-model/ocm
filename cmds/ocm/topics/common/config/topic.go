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

package topicconfig

import (
	"github.com/open-component-model/ocm/cmds/ocm/clictx"
	"github.com/spf13/cobra"
)

func New(ctx clictx.Context) *cobra.Command {
	return &cobra.Command{
		Use:   "configfile",
		Short: "configuration file",
		Example: `
type: generic.config.ocm.gardener.cloud/v1
configurations:
  - type: credentials.config.ocm.gardener.cloud
    repositories:
      - repository:
          type: DockerConfig/v1
          dockerConfigFile: "~/.docker/config.json"
          propagateConsumerIdentity: true
   - type: attributes.config.ocm.gardener.cloud
     attributes:  # map of attribute settings
       compat: true
#  - type: scripts.ocm.config.ocm.gardener.cloud
#    scripts:
#      "default":
#         script:
#           process: true
`,
		Long: `
The command line client supports configuring by a given configuration file.
If existent by default the file <code>$HOME/.ocmconfig</code> will be read.
Using the option <code>--config</code> an alternative file can be specified.

The file format is yaml. It uses the same type mechanism used for all
kinds of typed specification in the ocm area. The file must have the type of
a configuration specification. Instead, the command line client supports
a generic configuration specification able to host a list of arbitrary configuration
specifications. The type for this spec is <code>generic.config.ocm.gardener.cloud/v1</code>.
`,
	}
}
