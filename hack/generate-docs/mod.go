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
	"strings"

	"github.com/spf13/cobra"
)

// UseLine puts out the full usage for a given command (including parents).
func UseLine(c *cobra.Command) string {
	useline := c.Use
	if !strings.Contains(useline, " ") {
		// no syntax given
		if c.HasAvailableLocalFlags() {
			useline += " [<options>]"
		}
		if c.HasAvailableSubCommands() {
			if c.Runnable() {
				useline += " [<sub command> ...]"
			} else {
				useline += " <sub command> ..."
			}
		}
	}
	if c.HasParent() {
		useline = c.Parent().CommandPath() + " " + useline
	}
	if c.DisableFlagsInUseLine {
		return useline
	}
	if c.HasAvailableFlags() && !strings.Contains(useline, "[<options>]") {
		useline += " [<options>]"
	}
	return useline
}
