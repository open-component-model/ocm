package v1

type StringMap map[string]string

// Copy copies map.
func (l StringMap) Copy() StringMap {
	n := StringMap{}
	for k, v := range l {
		n[k] = v
	}
	return n
}
