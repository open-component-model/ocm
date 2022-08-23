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

package cobrautils

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func LinkForCmd(cmd *cobra.Command) string {
	return LinkForPath(cmd.CommandPath())
}

func LinkForPath(path string) string {
	link := path + ".md"
	link = strings.Replace(link, " ", "_", -1)
	return link
}

func FormatLink(pname string, linkhandler func(string) string) string {
	return fmt.Sprintf("[%s](%s)", pname, linkhandler((pname)))
}

func FormatLinkWithHandler(linkhandler func(string) string) func(string) string {
	return func(pname string) string {
		return FormatLink(pname, linkhandler)
	}
}

func SubstituteCommandLinks(desc string, linkformat func(string) string) (string, error) {
	for {
		link := strings.Index(desc, "<CMD>")
		if link < 0 {
			return desc, nil
		}
		end := strings.Index(desc, "</CMD>")
		if end < 0 {
			return "", fmt.Errorf("missing </CMD> in: %s\n" + desc)
		}
		path := desc[link+5 : end]
		desc = desc[:link] + linkformat(path) + desc[end+6:]
	}
}
