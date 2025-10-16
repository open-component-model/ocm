package version

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"

	"github.com/Masterminds/semver/v3"
	"ocm.software/ocm"
)

var (
	gitVersion   = "0.0.0-dev"
	gitCommit    string
	gitTreeState string
	buildDate    = "1970-01-01T00:00:00Z"
)

func init() {
	if gitVersion == "0.0.0-dev" {
		// gitVersion = strings.TrimSpace(string(MustAsset("../../VERSION")))
		gitVersion = strings.TrimSpace(ocm.Version)
	}
}

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

// String returns info as a short semantic version string (0.8.15).
func (info Info) SemVer() string {
	return info.Major + "." + info.Minor + "." + info.Patch
}

// String returns current Release version.
func Current() string {
	return Get().SemVer()
}

// Get returns the overall codebase version. It's for detecting
// what code a binary was built from.
// These variables typically come from -ldflags settings and in
// their absence fallback to the settings in pkg/version/base.go.
func Get() Info {
	var (
		gitMajor string
		gitMinor string
		gitPatch = "0"
		gitPre   string
		gitMeta  string
	)

	v, err := semver.NewVersion(gitVersion)
	if err == nil {
		gitMajor = strconv.FormatUint(v.Major(), 10)
		gitMinor = strconv.FormatUint(v.Minor(), 10)
		gitPatch = strconv.FormatUint(v.Patch(), 10)
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
