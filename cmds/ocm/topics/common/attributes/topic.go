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

package attributes

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/open-component-model/ocm/pkg/contexts/clictx"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext"
)

func New(ctx clictx.Context) *cobra.Command {
	return &cobra.Command{
		Use:   "attributes",
		Short: "configuration attributes used to control the behaviour",
		Long: `
The OCM library supports are set of attributes, which can be used to influence
the bahaviour of various functions. The CLI also supports setting of those
attributes using the config file (see <CMD>ocm configfile</CMD>) or by
command line options of the main command (see <CMD>ocm</CMD>).

The following options are available in the currently used version of the
OCM library:
` + Attributes(),
	}
}

func Attributes() string {
	s := ""
	sep := ""
	for _, a := range datacontext.DefaultAttributeScheme.KnownTypeNames() {
		t, _ := datacontext.DefaultAttributeScheme.GetType(a)
		desc := t.Description()
		if !strings.Contains(desc, "not via command line") {
			for strings.HasPrefix(desc, "\n") {
				desc = desc[1:]
			}
			for strings.HasSuffix(desc, "\n") {
				desc = desc[:len(desc)-1]
			}
			lines := strings.Split(desc, "\n")
			title := lines[0]
			desc = "  " + strings.Join(lines[1:], "\n  ")
			short := ""
			for k, v := range datacontext.DefaultAttributeScheme.Shortcuts() {
				if v == a {
					short = short + ",<code>" + k + "</code>"
				}
			}
			if len(short) > 0 {
				short = " [" + short[1:] + "]"
			}
			s = fmt.Sprintf("%s%s- <code>%s</code>%s: %s\n\n%s", s, sep, a, short, title, desc)
			sep = "\n\n"
		}
	}
	return s
}
