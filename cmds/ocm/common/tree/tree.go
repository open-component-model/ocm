package tree

import (
	"fmt"
	"strings"

	"ocm.software/ocm/api/utils"
	common "ocm.software/ocm/api/utils/misc"
)

type Object interface {
	common.HistorySource
	IsNode() *common.NameVersion
}

type Typed interface {
	GetKind() string
}

type ValidTreeElement interface {
	IsValid() bool
}

type NodeCreator func(common.History, common.NameVersion) Object

// TreeObject is an element enriched by a textual
// tree graph prefix line.
type TreeObject struct {
	Graph  string
	Object Object
	Node   *TreeNode // for synthesized nodes this entry is used if no object can be synthesized
}

func (t *TreeObject) String() string {
	if t.Object != nil {
		return fmt.Sprintf("%s %s", t.Graph, t.Object)
	}
	return fmt.Sprintf("%s %s", t.Graph, t.Node.String())
}

type TreeNode struct {
	common.NameVersion
	History  common.History
	CausedBy Object // the object causing the synthesized node to be inserted
}

var (
	vertical   = "│" + space[1:]
	horizontal = "─"
	corner     = "└" + horizontal
	fork       = "├" + horizontal
	space      = "   "
	node       = "⊗" // \u2297
)

// MapToTree maps a list of elements featuring a resolution history
// into a list of elements providing an ascii tree graph field
// Intermediate nodes are synthesized, so only leaf elements are required.
// If an element should act as explicit node, it must state to be a node,
// in this case the node will be tagged with the nodeSymbol. If this
// is not desired, pass an empty symbol string.
func MapToTree(objs Objects, creator NodeCreator, symbols ...string) TreeObjects {
	result := TreeObjects{}
	nodeSym := utils.OptionalDefaulted(node, symbols...)
	if nodeSym != "" && !strings.HasPrefix(nodeSym, " ") {
		nodeSym = " " + nodeSym
	}
	handleLevel(objs, "", nil, 0, creator, &result, nodeSym)
	return result
}

func handleLevel(objs Objects, header string, prefix common.History, start int, creator NodeCreator, result *TreeObjects, nodeSym string) {
	var node *common.NameVersion
	lvl := len(prefix)
	for i := start; i < len(objs); {
		var next int
		h := objs[i].GetHistory()
		if !h.HasPrefix(prefix) {
			return
		}
		ftag := corner
		stag := space
		key := objs[i].IsNode()
		for next = i + 1; next < len(objs); next++ {
			if s := objs[next].GetHistory(); s.HasPrefix(prefix) {
				if len(s) > lvl && len(h) > lvl && h[lvl] == s[lvl] { // skip same sub level
					continue
				}
				if key != nil {
					if len(s) > lvl && *key == s[lvl] { // skip same sub level
						continue
					}
				}
				ftag = fork
				stag = vertical
			}
			break
		}
		if len(h) == lvl {
			node = objs[i].IsNode() // Element acts as dedicate node
			sym := ""
			if node != nil {
				if i < len(objs)-1 {
					sub := objs[i+1].GetHistory()
					if len(sub) > len(h) && sub.HasPrefix(append(h, *node)) {
						sym = nodeSym
					}
				}
			}
			if t, ok := objs[i].(Typed); ok {
				k := t.GetKind()
				if k != "" {
					sym += " " + k
				}
			}
			if valid, ok := objs[i].(ValidTreeElement); !ok || valid.IsValid() {
				*result = append(*result, &TreeObject{
					Graph:  header + ftag + sym,
					Object: objs[i],
				})
			}
			i++
		} else {
			if node == nil || *node != h[lvl] {
				// synthesize node if only leafs or non-matching node has been issued before
				var o Object
				var n *TreeNode
				if creator != nil {
					o = creator(prefix, h[len(prefix)])
				}
				if o == nil {
					n = &TreeNode{h[len(prefix)], prefix, objs[i]}
				}
				*result = append(*result, &TreeObject{
					Graph:  header + ftag, // + " " + h[len(prefix)].String(),
					Object: o,
					Node:   n,
				})
			}
			handleLevel(objs, header+stag, h[:len(prefix)+1], i, creator, result, nodeSym)
			i = next
			node = nil
		}
	}
}
