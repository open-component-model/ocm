package utils

import (
	"fmt"
	"sort"
	"strings"

	"github.com/mandelsoft/goutils/maputils"
)

type DescriptionProvider interface {
	GetDescription() string
}

type KeyInfo interface {
	DescriptionProvider
	GetKey() string
}

func FormatKey(k string) string {
	return strings.ReplaceAll(k, "<", "&lt;")
}

func FormatList(def string, elems ...KeyInfo) string {
	names := ""
	for _, n := range elems {
		add := ""
		if n.GetKey() == def {
			add = " (default)"
		}
		names = fmt.Sprintf("%s\n  - <code>%s</code>:%s %s", names, FormatKey(n.GetKey()), add, n.GetDescription())
	}
	return names
}

func FormatMap[T DescriptionProvider](def string, elems map[string]T) string {
	keys := maputils.OrderedKeys(elems)
	sort.Strings(keys)
	names := ""
	for _, k := range keys {
		e := elems[k]
		add := ""
		if k == def {
			add = " (default)"
		}
		names = fmt.Sprintf("%s\n  - <code>%s</code>:%s %s", names, FormatKey(k), add, e.GetDescription())
	}
	return names
}
