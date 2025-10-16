package registrations

import (
	"sort"
	"strings"

	"github.com/mandelsoft/goutils/general"
	"ocm.software/ocm/api/utils/listformat"
)

type HandlerInfos []HandlerInfo

var (
	_ listformat.ListElements = HandlerInfos(nil)
	_ sort.Interface          = HandlerInfos(nil)
)

func (h HandlerInfos) Len() int {
	return len(h)
}

func (h HandlerInfos) Less(i, j int) bool {
	return strings.Compare(h[i].Name, h[j].Name) < 0
}

func (h HandlerInfos) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h HandlerInfos) Key(i int) string {
	return h[i].Name
}

func (h HandlerInfos) Description(i int) string {
	var desc string

	if h[i].Node {
		desc = "[" + general.Conditional(h[i].ShortDesc == "", "intermediate", strings.Trim(h[i].ShortDesc, "\n")) + "]"
	} else {
		desc = h[i].ShortDesc
	}
	return desc + general.Conditional(h[i].Description == "", "", "\n\n"+strings.Trim(h[i].Description, "\n"))
}

type HandlerInfo struct {
	Name        string
	ShortDesc   string
	Description string
	Node        bool
}

func NewLeafHandlerInfo(short, desc string) HandlerInfos {
	return HandlerInfos{
		{
			ShortDesc:   short,
			Description: desc,
		},
	}
}

func NewNodeHandlerInfo(short, desc string) HandlerInfos {
	return HandlerInfos{
		{
			ShortDesc:   short,
			Description: desc,
			Node:        true,
		},
	}
}
