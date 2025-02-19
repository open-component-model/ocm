package semverutils

import (
	"sort"
	"strings"

	"github.com/Masterminds/semver/v3"
	"slices"
)

type VersionCache map[string]*semver.Version

func (c VersionCache) Get(v string) (*semver.Version, error) {
	if s := c[v]; s != nil {
		return s, nil
	}
	s, err := semver.NewVersion(v)
	if err != nil {
		return nil, err
	}
	c[v] = s
	return s, nil
}

func (c VersionCache) Compare(a, b string) int {
	va, err := c.Get(a)
	if err != nil {
		return strings.Compare(a, b)
	}
	vb, err := c.Get(b)
	if err != nil {
		return strings.Compare(a, b)
	}
	return va.Compare(vb)
}

func Compare(a, b string) int {
	va, err := semver.NewVersion(a)
	if err != nil {
		return strings.Compare(a, b)
	}
	vb, err := semver.NewVersion(b)
	if err != nil {
		return strings.Compare(a, b)
	}
	return va.Compare(vb)
}

func SortVersions(vers []string) error {
	cache := VersionCache{}
	for _, v := range vers {
		_, err := cache.Get(v)
		if err != nil {
			return err
		}
	}

	sort.Slice(vers, func(a, b int) bool {
		va, _ := cache.Get(vers[a])
		vb, _ := cache.Get(vers[b])
		return va.Compare(vb) < 0
	})
	return nil
}

func Latest(vers []string) (string, error) {
	if len(vers) == 0 {
		return "", nil
	}
	vers = slices.Clone(vers)
	err := SortVersions(vers)
	if err != nil {
		return "", err
	}
	return vers[len(vers)-1], nil
}
