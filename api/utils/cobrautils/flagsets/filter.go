package flagsets

func Not(f Filter) Filter {
	return func(name string) bool {
		return !f(name)
	}
}

func And(fs ...Filter) Filter {
	return func(name string) bool {
		for _, f := range fs {
			if !f(name) {
				return false
			}
		}
		return true
	}
}

func Or(fs ...Filter) Filter {
	return func(name string) bool {
		for _, f := range fs {
			if f(name) {
				return true
			}
		}
		return false
	}
}

func Changed(opts ConfigOptions) Filter {
	return func(name string) bool {
		return opts.Changed(name)
	}
}
