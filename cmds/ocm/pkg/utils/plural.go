package utils

import (
	pluralize "github.com/gertd/go-pluralize"
	"golang.org/x/exp/constraints"
)

var client = pluralize.NewClient()

func Plural[N constraints.Integer](s string, amount ...N) string {
	var n N = 0
	for _, a := range amount {
		n += a
	}
	if n == 1 {
		return s
	}
	return client.Plural(s)
}
