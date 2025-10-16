package common

import (
	"fmt"
	"strings"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/general"
	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/compdesc"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	utils2 "ocm.software/ocm/api/utils"
	"ocm.software/ocm/cmds/ocm/common/options"
)

func ConsumeIdentities(pattern bool, args []string, stop ...string) ([]metav1.Identity, []string, error) {
	result := []metav1.Identity{}
	for i, a := range args {
		for _, s := range stop {
			if s == a {
				return result, args[i+1:], nil
			}
		}
		i := strings.Index(a, "=")
		if i < 0 {
			result = append(result, metav1.Identity{compdesc.SystemIdentityName: a})
		} else {
			if len(result) == 0 {
				if !pattern {
					return nil, nil, fmt.Errorf("first resource identity argument must be a sole resource name")
				}
				result = append(result, metav1.Identity{a[:i]: a[i+1:]})
			} else {
				result[len(result)-1][a[:i]] = a[i+1:]
			}
			if i == 0 {
				return nil, nil, fmt.Errorf("extra identity key might not be empty in %q", a)
			}
		}
	}
	return result, nil, nil
}

func MapArgsToIdentities(args ...string) ([]metav1.Identity, error) {
	result, _, err := ConsumeIdentities(false, args)
	return result, err
}

func MapArgsToIdentityPattern(args ...string) (metav1.Identity, error) {
	result, _, err := ConsumeIdentities(true, args)
	if err == nil {
		switch len(result) {
		case 0:
			return nil, nil
		case 1:
			return result[0], err
		default:
			if len(result) > 1 {
				return nil, errors.Newf("only one identity pattern possible (sole name in between)")
			}
		}
	}
	return nil, err
}

////////////////////////////////////////////////////////////////////////////////

// OptionWithSessionCompleter describes the interface for option objects requiring
// a completion with a session.
type OptionWithSessionCompleter interface {
	CompleteWithSession(ctx clictx.OCM, session ocm.Session) error
}

// CompleteOptionsWithSession provides an options.OptionsProcessor completing
// options by passing a session object using the OptionWithSessionCompleter interface.
// If an optional argument true is given, it also tries the other standard completion
// methods possible for an options object.
func CompleteOptionsWithSession(ctx clictx.Context, session ocm.Session, all ...bool) options.OptionsProcessor {
	otherCompleters := general.Optional(all...)
	return func(opt options.Options) error {
		if c, ok := opt.(OptionWithSessionCompleter); ok {
			return c.CompleteWithSession(ctx.OCM(), session)
		}
		if otherCompleters {
			if c, ok := opt.(options.OptionWithCLIContextCompleter); ok {
				return c.Configure(ctx)
			}
			if c, ok := opt.(options.SimpleOptionCompleter); ok {
				return c.Complete()
			}
		}
		return nil
	}
}

////////////////////////////////////////////////////////////////////////////////

func MapLabelSpecs(d interface{}) (interface{}, error) {
	if d == nil {
		return nil, nil
	}

	m, ok := d.(map[string]interface{})
	if !ok {
		return nil, errors.ErrInvalid("go type", fmt.Sprintf("%T", d))
	}

	var labels []interface{}
	found := map[string]struct{}{}
	for _, k := range utils2.StringMapKeys(m) {
		v := m[k]
		entry := map[string]interface{}{}
		if strings.HasPrefix(k, "*") {
			entry["signing"] = true
			k = k[1:]
		}
		if i := strings.Index(k, "@"); i > 0 {
			vers := k[i+1:]
			if !metav1.CheckLabelVersion(vers) {
				return nil, errors.ErrInvalid("invalid version %q for label %q", vers, k[:i])
			}
			entry["version"] = vers
			k = k[:i]
		}
		if _, ok := found[k]; ok {
			return nil, fmt.Errorf("duplicate label %q", k)
		}
		entry["name"] = k
		entry["value"] = v
		labels = append(labels, entry)
	}
	return labels, nil
}
