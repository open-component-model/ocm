package output

import (
	"strings"
)

func SelectBest(name string, candidates ...string) (string, int) {
	for i, c := range candidates {
		if strings.EqualFold(name, c) {
			return c, i
		}
	}
	return "", -1
}
