package data

func Slice(s Iterable) []interface{} {
	var a []interface{}
	i := s.Iterator()
	for i.HasNext() {
		a = append(a, i.Next())
	}
	return a
}

func StringArraySlice(s Iterable) [][]string {
	a := [][]string{}
	i := s.Iterator()
	for i.HasNext() {
		a = append(a, i.Next().([]string))
	}
	return a
}
