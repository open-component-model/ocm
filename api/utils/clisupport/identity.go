package clisupport

import (
	"fmt"
	"strings"

	"github.com/mandelsoft/goutils/errors"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
)

func ParseIdentityPath(ids ...string) ([]metav1.Identity, error) {
	var err error
	result := []metav1.Identity{}

	var id metav1.Identity
	for _, l := range ids {
		name, value, err := ParseIdentityAttribute(l)
		if err != nil {
			return nil, err
		}
		if name != "name" {
			if id == nil {
				return nil, fmt.Errorf("first attribute must be the name attribute")
			}
			if id[name] != "" {
				return nil, fmt.Errorf("attribute %q already set", name)
			}
			id[name] = value
		} else {
			if id != nil {
				result = append(result, id)
			}
			id = metav1.Identity{name: value}
		}
	}
	result = append(result, id)
	return result, err
}

func ParseIdentityAttribute(a string) (string, string, error) {
	i := strings.Index(a, "=")
	if i < 0 {
		return "", "", errors.ErrInvalid("identity attribute", a)
	}
	name := a[:i]
	value := a[i+1:]

	return name, value, nil
}
