package substutils

import (
	"github.com/drone/envsubst"
)

// Substitution maps a variable to a value.
// It reports whether a substitution for the given variable is
// available.
// if there is no substitution supported for the given variable
// and empty string should be returned.
type Substitution interface {
	Substitute(variable string) (string, bool)
}

type SubstitutionFunc func(variable string) (string, bool)

func (f SubstitutionFunc) Substitute(variable string) (string, bool) {
	return f(variable)
}

// Join combines any numer of Substitution objects.
// It provides the *first* valid substitution.
func Join(subst ...Substitution) Substitution {
	return SubstitutionFunc(func(variable string) (string, bool) {
		for _, s := range subst {
			r, ok := s.Substitute(variable)
			if ok {
				return r, true
			}
		}
		return "", false
	})
}

// Override combines any number of Substitution objects.
// It provides the *last* valid substitution.
func Override(subst ...Substitution) Substitution {
	return SubstitutionFunc(func(variable string) (string, bool) {
		for i := range len(subst) {
			r, ok := subst[len(subst)-i-1].Substitute(variable)
			if ok {
				return r, true
			}
		}
		return "", false
	})
}

func Eval(in string, subst ...Substitution) (string, error) {
	return envsubst.Eval(in, func(variable string) string {
		r, _ := Override(subst...).Substitute(variable)
		return r
	})
}
