// Package substutils provides some useful substitution
// functions for [envsubst.Eval] by offering
// a [Substitution] interface, some standard
// implementations and composition functions.
package substutils

import (
	"github.com/drone/envsubst"
)

var _ = envsubst.Eval
