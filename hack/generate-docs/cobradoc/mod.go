// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package cobradoc

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
