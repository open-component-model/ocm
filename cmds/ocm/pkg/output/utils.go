package output

// Fields composes a (string) field list based on a sequence of strings and or
// field lists.
func Fields(fields ...interface{}) []string {
	var result []string
	for _, f := range fields {
		switch v := f.(type) {
		case string:
			result = append(result, v)
		case []string:
			result = append(result, v...)
		case []interface{}:
			result = append(result, Fields(v...)...)
		}
	}
	return result
}
