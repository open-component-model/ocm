package common

import (
	"fmt"
	"strings"
)

type Stringer interface {
	String() string
}

type Element[C Stringer] interface {
	GetName() string
	GetConstraints() []C
}

func DescribeElements[E Element[C], C Stringer](elems []E) string {
	var list []string
	for _, m := range elems {
		n := m.GetName()
		var clist []string
		for _, c := range m.GetConstraints() {
			clist = append(clist, c.String())
		}
		if len(clist) > 0 {
			n = fmt.Sprintf("%s[%s]", n, strings.Join(clist, ","))
		}
		list = append(list, n)
	}
	return strings.Join(list, ",")
}
