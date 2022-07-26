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

package version

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"

	"github.com/Masterminds/semver/v3"
)

var (
	gitVersion   = "0.0.0-dev"
	gitCommit    string
	gitTreeState string
	buildDate    = "1970-01-01T00:00:00Z"
)

type Info struct {
	Major        string `json:"major"`
	Minor        string `json:"minor"`
	Patch        string `json:"patch"`
	PreRelease   string `json:"prerelease"`
	Meta         string `json:"meta"`
	GitVersion   string `json:"gitVersion"`
	GitCommit    string `json:"gitCommit"`
	GitTreeState string `json:"gitTreeState"`
	BuildDate    string `json:"buildDate"`
	GoVersion    string `json:"goVersion"`
	Compiler     string `json:"compiler"`
	Platform     string `json:"platform"`
}

// String returns info as a human-friendly version string.
func (info Info) String() string {
	return info.GitVersion
}

// GetInterface returns the overall codebase version. It's for detecting
// what code a binary was built from.
// These variables typically come from -ldflags settings and in
// their absence fallback to the settings in pkg/version/base.go
func Get() Info {
	var (
		gitMajor string
		gitMinor string
		gitPatch string = "0"
		gitPre   string
		gitMeta  string
	)

	v, err := semver.NewVersion(gitVersion)
	if err == nil {
		gitMajor = strconv.Itoa(int(v.Major()))
		gitMinor = strconv.Itoa(int(v.Minor()))
		gitPatch = strconv.Itoa(int(v.Patch()))
		gitPre = v.Prerelease()
		gitMeta = v.Metadata()
	} else {
		version := gitVersion
		if i := strings.Index(version, "-"); i >= 0 {
			gitPre = version[i+1:]
			version = version[:i]
		}
		if i := strings.Index(version, "+"); i >= 0 {
			gitMeta = version[i+1:]
			version = version[:i]
		}
		if i := strings.Index(gitPre, "+"); i >= 0 {
			gitMeta = gitPre[i+1:]
			gitPre = gitPre[:i]
		}
		v := strings.Split(version, ".")
		if len(v) >= 2 {
			gitMajor = v[0]
			gitMinor = v[1]
			if len(v) >= 3 {
				gitPatch = v[2]
			}
		}
	}

	return Info{
		Major:        gitMajor,
		Minor:        gitMinor,
		Patch:        gitPatch,
		PreRelease:   gitPre,
		Meta:         gitMeta,
		GitVersion:   gitVersion,
		GitCommit:    gitCommit,
		GitTreeState: gitTreeState,
		BuildDate:    buildDate,
		GoVersion:    runtime.Version(),
		Compiler:     runtime.Compiler,
		Platform:     fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}
