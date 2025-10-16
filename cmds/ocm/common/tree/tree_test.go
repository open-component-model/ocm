package tree_test

import (
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	common "ocm.software/ocm/api/utils/misc"
	"ocm.software/ocm/cmds/ocm/common/tree"
)

type Elem struct {
	common.History
	Node bool
	Data string
}

var _ tree.Object = (*Elem)(nil)

func (e *Elem) GetHistory() common.History {
	return e.History
}

func (e *Elem) IsNode() *common.NameVersion {
	if e.Node {
		n := common.NewNameVersion(e.Data, "")
		return &n
	}
	return nil
}

func (e *Elem) String() string {
	return e.Data
}

type Invalid struct {
	Elem
}

func (e *Invalid) IsValid() bool {
	return false
}

func I(hist ...string) *Invalid {
	h := common.History{}
	for _, v := range hist {
		h = append(h, common.NewNameVersion(v, ""))
	}
	return &Invalid{Elem{h, false, ""}}
}

func E(d string, hist ...string) *Elem {
	h := common.History{}
	for _, v := range hist {
		h = append(h, common.NewNameVersion(v, ""))
	}
	return &Elem{h, false, d}
}

func N(d string, hist ...string) *Elem {
	h := common.History{}
	for _, v := range hist {
		h = append(h, common.NewNameVersion(v, ""))
	}
	return &Elem{h, true, d}
}

func Check(t tree.TreeObjects, result string) {
	lines := strings.Split(result, "\n")
	for len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	for len(lines) > 0 && lines[0] == "" {
		lines = lines[1:]
	}
	min := len(t)
	if len(lines) < min {
		min = len(lines)
	}

	for i, l := range lines {
		if i >= min {
			fmt.Println(t)
			Fail(fmt.Sprintf("additional %d lines expected (%s...)", len(lines)-min, l), 1)
		}
		if l != t[i].String() {
			fmt.Println(t)
			Fail(fmt.Sprintf("mismatch of line %d:\nfound:    %s\nexpected: %s\n", i+1, t[i].String(), l), 1)
		}
	}
	if min < len(t) {
		fmt.Println(t)
		Fail(fmt.Sprintf("additional %d lines found in output (%s...)", len(t)-min, t[min]), 1)
	}
}

var _ = Describe("tree", func() {
	It("composes flat tree", func() {
		data := []tree.Object{
			E("a"),
			E("b"),
			E("c"),
		}

		t := tree.MapToTree(data, nil)
		Check(t, `
├─ a
├─ b
└─ c
`)
	})
	It("composes simple tree with nested elements", func() {
		data := []tree.Object{
			E("a"),
			E("a", "b"),
			E("a", "c"),
		}

		t := tree.MapToTree(data, nil)
		Check(t, `
├─ a
├─ b
│  └─ a
└─ c
   └─ a
`)
	})

	It("composes simple tree with nested elements", func() {
		data := []tree.Object{
			E("a"),
			E("b"),
			E("a", "b"),
			E("a", "c"),
		}

		t := tree.MapToTree(data, nil)
		Check(t, `
├─ a
├─ b
├─ b
│  └─ a
└─ c
   └─ a
`)
	})

	It("composes simple node tree with nested elements", func() {
		data := []tree.Object{
			N("a"),
			N("b"),
			N("a", "b"),
			N("c"),
			N("a", "c"),
		}

		t := tree.MapToTree(data, nil)
		Check(t, `
├─ a
├─ ⊗ b
│  └─ a
└─ ⊗ c
   └─ a
`)
	})
	It("composes simple node tree with nested elements as untagged nodes", func() {
		data := []tree.Object{
			N("a"),
			N("b"),
			N("a", "b"),
			N("c"),
			N("a", "c"),
		}

		t := tree.MapToTree(data, nil, "")
		Check(t, `
├─ a
├─ b
│  └─ a
└─ c
   └─ a
`)
	})

	It("some complex stuff", func() {
		data := []tree.Object{
			E("a"),
			N("b"),
			E("a", "b"),
			E("a", "b", "c"),
			E("a", "e", "f"),
			E("c"),
			E("d"),
		}

		t := tree.MapToTree(data, nil)
		Check(t, `
├─ a
├─ ⊗ b
│  ├─ a
│  └─ c
│     └─ a
├─ e
│  └─ f
│     └─ a
├─ c
└─ d

`)
	})
	It("end/end/end", func() {
		data := []tree.Object{
			N("b", "a"),
			N("c", "a", "b"),
			N("d", "a", "b"),
		}

		t := tree.MapToTree(data, nil)
		Check(t, `
└─ a
   └─ ⊗ b
      ├─ c
      └─ d
`)
	})
	It("endintermediate/end", func() {
		data := []tree.Object{
			N("b", "a"),
			N("c", "a", "b"),
			N("d", "a", "b"),
			N("e", "a"),
		}

		t := tree.MapToTree(data, nil)
		Check(t, `
└─ a
   ├─ ⊗ b
   │  ├─ c
   │  └─ d
   └─ e
`)
	})

	It("endintermediate/end", func() {
		data := []tree.Object{
			N("d6c3"),
			N("439d", "d6c3"),
			N("2c3e", "d6c3"),
			N("efbf", "d6c3", "2c3e"),
			N("60b2", "d6c3"),
		}

		t := tree.MapToTree(data, nil)
		Check(t, `
└─ ⊗ d6c3
   ├─ 439d
   ├─ ⊗ 2c3e
   │  └─ efbf
   └─ 60b2
`)
	})

	It("synthesizes empty node", func() {
		data := []tree.Object{
			I("b"),
		}

		t := tree.MapToTree(data, nil)
		Check(t, `
└─ b
`)
	})

	It("synthesizes intermediate empty node", func() {
		data := []tree.Object{
			E("a", "a"),
			E("b", "a"),
			I("b"),
			E("a", "c", "a"),
		}

		t := tree.MapToTree(data, nil)
		Check(t, `
├─ a
│  ├─ a
│  └─ b
├─ b
└─ c
   └─ a
      └─ a
`)
	})
})
