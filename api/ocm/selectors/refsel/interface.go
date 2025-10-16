package refsel

import (
	"regexp"

	"github.com/gobwas/glob"
	"ocm.software/ocm/api/ocm/selectors"
	"ocm.software/ocm/api/ocm/selectors/accessors"
)

type (
	Selector     = selectors.ReferenceSelector
	SelectorFunc = selectors.ReferenceSelectorFunc
)

////////////////////////////////////////////////////////////////////////////////

type Component string

func (c Component) MatchReference(list accessors.ElementListAccessor, ref accessors.ReferenceAccessor) bool {
	return string(c) == ref.GetComponentName()
}

////////////////////////////////////////////////////////////////////////////////

type compGlob struct {
	glob.Glob
}

func (c *compGlob) MatchReference(list accessors.ElementListAccessor, ref accessors.ReferenceAccessor) bool {
	if c.Glob == nil {
		return false
	}
	return c.Glob.Match(ref.GetComponentName())
}

func ComponentGlob(g string) Selector {
	c, err := glob.Compile(g, '/')
	return selectors.NewReferenceErrorSelectorImpl(&compGlob{c}, err)
}

////////////////////////////////////////////////////////////////////////////////

type compRegEx struct {
	*regexp.Regexp
}

func (c *compRegEx) MatchReference(list accessors.ElementListAccessor, ref accessors.ReferenceAccessor) bool {
	if c.Regexp == nil {
		return false
	}
	return c.Regexp.MatchString(ref.GetComponentName())
}

func ComponentRegex(g string) selectors.ReferenceSelector {
	c, err := regexp.Compile(g)
	return selectors.NewReferenceErrorSelectorImpl(&compRegEx{c}, err)
}
