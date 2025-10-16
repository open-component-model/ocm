package options

import (
	"reflect"

	"github.com/mandelsoft/goutils/generics"
	"github.com/spf13/pflag"
	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/api/utils/out"
)

// OptionsProcessor is handler used to process all
// option found in a set of options.
type OptionsProcessor func(Options) error

// SimpleOptionCompleter describes the interface for an option object
// requirung completion without any further information.
type SimpleOptionCompleter interface {
	Complete() error
}

// OptionWithOutputContextCompleter describes the interface for an option object
// requirung completion with an output context.
type OptionWithOutputContextCompleter interface {
	Complete(ctx out.Context) error
}

// OptionWithCLIContextCompleter describes the interface for an option object
// requirung completion with a CLI context.
type OptionWithCLIContextCompleter interface {
	Configure(ctx clictx.Context) error
}

type Usage interface {
	Usage() string
}

type Options interface {
	AddFlags(fs *pflag.FlagSet)
}

////////////////////////////////////////////////////////////////////////////////

type OptionSelector func(provider OptionSetProvider) bool

func Not(s OptionSelector) OptionSelector {
	return func(provider OptionSetProvider) bool {
		return !s(provider)
	}
}

func Always() OptionSelector {
	return func(provider OptionSetProvider) bool {
		return true
	}
}

func Never() OptionSelector {
	return func(provider OptionSetProvider) bool {
		return false
	}
}

////////////////////////////////////////////////////////////////////////////////

type OptionSet []Options

type OptionSetProvider interface {
	AsOptionSet() OptionSet
}

func (s OptionSet) AddFlags(fs *pflag.FlagSet) {
	for _, o := range s {
		o.AddFlags(fs)
	}
}

func (s OptionSet) AsOptionSet() OptionSet {
	return s
}

func (s OptionSet) Usage() string {
	u := ""
	for _, n := range s {
		if c, ok := n.(Usage); ok {
			u += "\n" + c.Usage()
		}
	}
	return u
}

func (s OptionSet) Options(proto Options) interface{} {
	t := reflect.TypeOf(proto)
	for _, o := range s {
		if reflect.TypeOf(o) == t {
			return o
		}
		if set, ok := o.(OptionSetProvider); ok {
			r := set.AsOptionSet().Options(proto)
			if r != nil {
				return r
			}
		}
	}
	return nil
}

func FindOptions[T any](s OptionSetProvider) []T {
	var found []T

	t := generics.TypeOf[T]()
	for _, o := range s.AsOptionSet() {
		if reflect.TypeOf(o).AssignableTo(t) {
			found = append(found, generics.Cast[T](o))
		}
		if set, ok := o.(OptionSetProvider); ok {
			found = append(found, FindOptions[T](set)...)
		}
	}
	return found
}

// Get extracts the option for a given target. This might be a
//   - pointer to a struct implementing the Options interface which
//     will fill the struct with a copy of the options OR
//   - a pointer to such a pointer which will be filled with the
//     pointer to the actual member of the OptionSet.
func (s OptionSet) Get(proto interface{}) bool {
	val := true
	t := reflect.TypeOf(proto)
	if t.Elem().Kind() == reflect.Ptr {
		t = t.Elem()
		val = false
	}
	for _, o := range s {
		if reflect.TypeOf(o) == t {
			if val {
				reflect.ValueOf(proto).Elem().Set(reflect.ValueOf(o).Elem())
			} else {
				reflect.ValueOf(proto).Elem().Set(reflect.ValueOf(o))
			}
			return true
		}
		if set, ok := o.(OptionSetProvider); ok {
			r := set.AsOptionSet().Get(proto)
			if r {
				return r
			}
		}
	}
	return false
}

// ProcessOnOptions processes all options found in the option set
// with a given OptionsProcessor.
func (s OptionSet) ProcessOnOptions(f OptionsProcessor) error {
	for _, n := range s {
		var err error
		if set, ok := n.(OptionSetProvider); ok {
			err = set.AsOptionSet().ProcessOnOptions(f)
			if err != nil {
				return err
			}
		}
		err = f(n)
		if err != nil {
			return err
		}
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////

func CompleteOptions(opt Options) error {
	if c, ok := opt.(SimpleOptionCompleter); ok {
		return c.Complete()
	}
	return nil
}

func CompleteOptionsWithCLIContext(ctx clictx.Context) OptionsProcessor {
	return func(opt Options) error {
		if c, ok := opt.(OptionWithCLIContextCompleter); ok {
			return c.Configure(ctx)
		}
		if c, ok := opt.(OptionWithOutputContextCompleter); ok {
			return c.Complete(ctx)
		}
		return CompleteOptions(opt)
	}
}

////////////////////////////////////////////////////////////////////////////////

type Condition interface {
	IsTrue() bool
}

type ConditionFunction func() bool

func (f ConditionFunction) IsTrue() bool {
	return f()
}

type Flag bool

func (f Flag) IsTrue() bool {
	return bool(f)
}

func Or(conditions ...Condition) Condition {
	return ConditionFunction(func() bool {
		for _, c := range conditions {
			if c.IsTrue() {
				return true
			}
		}
		return false
	})
}

func And(conditions ...Condition) Condition {
	return ConditionFunction(func() bool {
		for _, c := range conditions {
			if !c.IsTrue() {
				return false
			}
		}
		return true
	})
}
